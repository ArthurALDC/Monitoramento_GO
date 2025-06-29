// main.go
package main

import ( // OLÁ RAPAZES ESSES SÃO OS PACOTES QUE VAMOS USAR
	"fmt"  // pacote pra textos
	"log"  // pacote pra registrar as mensagens de erro
	"time" // para controlar o tempo

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

func getCpuUsage() {

	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Erro ao obter uso da CPU: %v", err)
		return
	}

	// 'percent' é um slice, mas com 'percpu=false', ele terá apenas um elemento.
	if len(percent) > 0 {
		fmt.Printf("Uso de CPU: %.2f%%\n", percent[0])
	}
}

func getRamUsage() {
	vmStat, err := mem.VirtualMemory() // vmstat da umas informações da memoria
	if err != nil {
		log.Printf("Erro ao obter uso de RAM: %v", err)
		return
	}
	// tudo ta em byte então converte para gb pra uma melgor leitura
	totalGB := float64(vmStat.Total) / (1024 * 1024 * 1024)
	usedGB := float64(vmStat.Used) / (1024 * 1024 * 1024)

	fmt.Printf(
		"Uso de RAM: %.2f GB / %.2f GB (%.2f%%)\n",
		usedGB,
		totalGB,
		vmStat.UsedPercent,
	)
}

func getDiskUsage() {

	// a biblioteca é pica ela vai pegar o caminho do sistema ta
	path := "/"
	diskStat, err := disk.Usage(path)
	if err != nil {
		log.Printf("Erro ao obter uso de Disco: %v", err)
		return
	}

	// conversao de novo
	totalGB := float64(diskStat.Total) / (1024 * 1024 * 1024)
	usedGB := float64(diskStat.Used) / (1024 * 1024 * 1024)

	fmt.Printf(
		"Uso de Disco (%s): %.2f GB / %.2f GB (%.2f%%)\n",
		path,
		usedGB,
		totalGB,
		diskStat.UsedPercent,
	)
}

// getNetUsage mostra ai os byte
func getNetUsage() {
	// 'pernic=false' soma as estatísticas de todas as interfaces
	ioCounters, err := net.IOCounters(false)
	if err != nil {
		log.Printf("Erro ao obter uso de Rede: %v", err)
		return
	}

	// ioCounters é um slice, mas com 'pernic=false', ele terá apenas um elemento, ou seja ela vai juntar
	// as informações das placas de redes e botar tudo num lugar só
	if len(ioCounters) > 0 {
		stats := ioCounters[0]
		// byte pra mb agora
		bytesSentMB := float64(stats.BytesSent) / (1024 * 1024)
		bytesRecvMB := float64(stats.BytesRecv) / (1024 * 1024)

		fmt.Printf(
			"Uso de Rede: %.2f MB enviados / %.2f MB recebidos\n",
			bytesSentMB,
			bytesRecvMB,
		)
	}
}

func main() {
	fmt.Println("--- Agente de Monitoramento de Recursos ---")
	fmt.Println("Pressione CTRL+C para sair.")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // Garante que o ticker seja limpo ao sair.

	for {

		<-ticker.C // é um canal usado para receber os dados

		fmt.Println("-------------------------------------------")
		fmt.Printf("Status em: %s\n", time.Now().Format("15:04:05")) // isso aqui um bizu que o gpt deu KKKKKKKKK ele falou que em GO
		// a formatação é usado com exemplos, achei tendencia

		getCpuUsage()
		getRamUsage()
		getDiskUsage()
		getNetUsage()

		fmt.Println("-------------------------------------------")
	}
}
