// main.go
package main

import ( // OLÁ RAPAZES ESSES SÃO OS PACOTES QUE VAMOS USAR
	"fmt" // pacote pra textos
	"image/color"
	"log" // pacote pra registrar as mensagens de erro
	"os"
	"runtime" // Importado para obter o número de núcleos da CPU
	"sort"

	// para controlar o tempo

	// Pacotes do Gio para criar a interface gráfica
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"

	// Pacotes do gopsutil para puxar os dados
	"github.com/shirou/gopsutil/v3/process"
)

// Lista com os nomes das abas
var nomesDasAbas = []string{"Processos", "Desempenho", "Historico", "Empty"}

// abaSelecionada indica qual aba está ativa
var abaSelecionada int
var ordenarPor = "CPU"       // opções: "CPU", "MEM", "ALFABETICO"
var ordenarCrescente = false // false = maior para menor, true = menor para maior
var btnCPU, btnMEM, btnALF, btnOrdem widget.Clickable

// botoesDasAbas são os botões clicáveis no topo da interface
var botoesDasAbas [4]widget.Clickable

func main() {
	// Inicia a janela da aplicação em uma goroutine
	go func() {
		janela := app.NewWindow()

		// Chama a função que desenha e controla a janela
		if err := iniciarLoopDaJanela(janela); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	// Inicia o loop principal do Gio
	app.Main()
}

// iniciarLoopDaJanela é o loop principal da UI
func iniciarLoopDaJanela(janela *app.Window) error {
	tema := material.NewTheme(gofont.Collection())
	var operacoes op.Ops

	for {
		evento := <-janela.Events()

		switch evento := evento.(type) {
		case system.DestroyEvent:
			return evento.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&operacoes, evento)

			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					var elementos []layout.FlexChild

					for i := range botoesDasAbas {
						botao := material.Button(tema, &botoesDasAbas[i], nomesDasAbas[i])
						idx := i
						elementos = append(elementos, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if botoesDasAbas[idx].Clicked() {
								abaSelecionada = idx
							}
							return botao.Layout(gtx)
						}))
					}

					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, elementos...)
				}),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if abaSelecionada == 0 {
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceAround,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(tema, &btnCPU, "CPU")
								if btnCPU.Clicked() {
									ordenarPor = "CPU"
								}
								return btn.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(tema, &btnMEM, "MEM")
								if btnMEM.Clicked() {
									ordenarPor = "MEM"
								}
								return btn.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(tema, &btnALF, "Alfabético")
								if btnALF.Clicked() {
									ordenarPor = "ALFABETICO"
								}
								return btn.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								textoBtn := "Decrescente"
								if ordenarCrescente {
									textoBtn = "Crescente"
								}
								btn := material.Button(tema, &btnOrdem, textoBtn)
								if btnOrdem.Clicked() {
									ordenarCrescente = !ordenarCrescente
								}
								return btn.Layout(gtx)
							}),
						)
					}
					return layout.Dimensions{}
				}),

				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					switch abaSelecionada {
					case 0:
						return desenharAbaProcessos(gtx, tema)
					default:
						texto := material.Body1(tema, fmt.Sprintf("Conteúdo da aba: %s (em branco)", nomesDasAbas[abaSelecionada]))
						texto.Color = color.NRGBA{A: 255}
						return texto.Layout(gtx)
					}
				}),
			)

			evento.Frame(gtx.Ops)
		}
	}
}

// Criar um widget.Clickable reutilizável para os botões de ordenação
func newClickable(id string) *widget.Clickable {
	// Armazenar por id não é necessário, só criar novo
	return &widget.Clickable{}
}

// substituir desenharAbaProcessos por essa versão:

func desenharAbaProcessos(gtx layout.Context, tema *material.Theme) layout.Dimensions {
	numCPU := runtime.NumCPU() * 2
	if numCPU == 0 {
		numCPU = 1
	}
	floatNumCPU := float64(numCPU)

	listaDeProcessos, err := process.Processes()
	if err != nil {
		log.Printf("Erro ao obter lista de processos: %v", err)
		return layout.Dimensions{}
	}

	type somaInfo struct {
		usoCPU float64
		usoRAM float32
		pids   map[int32]bool
	}

	processosMap := make(map[string]*somaInfo)

	for _, proc := range listaDeProcessos {
		nome, _ := proc.Name()
		if nome == "" {
			continue
		}

		pid := proc.Pid
		usoCPU, errCPU := proc.CPUPercent()
		usoRAM, errRAM := proc.MemoryPercent()
		if errCPU != nil || errRAM != nil {
			continue
		}

		if info, ok := processosMap[nome]; ok {
			if !info.pids[pid] {
				info.usoCPU += usoCPU
				info.usoRAM += usoRAM
				info.pids[pid] = true
			}
		} else {
			processosMap[nome] = &somaInfo{
				usoCPU: usoCPU,
				usoRAM: usoRAM,
				pids:   map[int32]bool{pid: true},
			}
		}
	}

	tipos := make([]string, 0, len(processosMap))
	for nome := range processosMap {
		tipos = append(tipos, nome)
	}

	switch ordenarPor {
	case "CPU":
		sort.Slice(tipos, func(i, j int) bool {
			if ordenarCrescente {
				return processosMap[tipos[i]].usoCPU < processosMap[tipos[j]].usoCPU
			}
			return processosMap[tipos[i]].usoCPU > processosMap[tipos[j]].usoCPU
		})
	case "MEM":
		sort.Slice(tipos, func(i, j int) bool {
			if ordenarCrescente {
				return processosMap[tipos[i]].usoRAM < processosMap[tipos[j]].usoRAM
			}
			return processosMap[tipos[i]].usoRAM > processosMap[tipos[j]].usoRAM
		})
	case "ALFABETICO":
		sort.Slice(tipos, func(i, j int) bool {
			if ordenarCrescente {
				return tipos[i] < tipos[j]
			}
			return tipos[i] > tipos[j]
		})
	}

	lista := layout.List{Axis: layout.Vertical}

	return lista.Layout(gtx, len(tipos), func(gtx layout.Context, i int) layout.Dimensions {
		if i >= len(tipos) {
			return layout.Dimensions{}
		}

		nome := tipos[i]
		info := processosMap[nome]
		usoCPUNormalizado := info.usoCPU / floatNumCPU

		texto := fmt.Sprintf("Nome: %-30s | CPU: %6.2f%% | MEM: %5.2f%% | Instâncias: %d",
			nome, usoCPUNormalizado, info.usoRAM, len(info.pids))

		return material.Body2(tema, texto).Layout(gtx)
	})
}
