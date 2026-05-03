package track

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func TrackUsage() {
	go trackCpuUsage()
}

// Fonction pour obtenir l'utilisation actuelle du CPU.
func getCpuUsage() (float64, error) {
	p, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	return p[0], nil
}

// Fonction pour calculer la moyenne d'un slice d'integers.
func mean(nums []uint64) uint64 {
	total := uint64(0)
	for _, num := range nums {
		total += num
	}
	return total / uint64(len(nums))
}

// Goroutine pour tenir à jour un historique d'utilisation du CPU sur les 10 dernières secondes.
func trackCpuUsage() {
	cpuUsages := make(chan float64)
	go func() {
		defer close(cpuUsages)
		for {
			usage, err := getCpuUsage()
			if err != nil {
				fmt.Println("Error getting CPU usage:", err)
			} else {
				cpuUsages <- usage
			}
			time.Sleep(time.Second)
		}
	}()
	memUsages := make([]uint64, 0, 30)
	for usage := range cpuUsages {
		if len(memUsages) > 30 {
			memUsages = memUsages[1:] // Si nous avons plus de 5 éléments, enlever le plus ancien.
		}
		memUsages = append(memUsages, uint64(usage))
		fmt.Printf("L'utilisation du CPU est de: %.2d%%\n", mean(memUsages))
	}
}
