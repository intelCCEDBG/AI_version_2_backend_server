package cleaning

import (
	"fmt"
	"os"
	"time"
)

func Clean_service(cycle int, period int) {
	for {
		currenttime := time.Now()
		cleaning(currenttime, period)
		time.Sleep(time.Duration(cycle) * time.Second)
	}
}

func cleaning(currenttime time.Time, period int) {
	items, _ := os.ReadDir("/home/media/video/")
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := os.ReadDir("/home/media/video/" + item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					// fmt.Println(subitem.Name())
					if len(subitem.Name()) > 19 {
						tm, err := time.Parse("2006-01-02_15-04-05", subitem.Name()[:19])
						if err != nil {
							fmt.Println(err)
						}
						if currenttime.Sub(tm).Hours() > float64(period) && err == nil {
							os.Remove("/home/media/video/" + item.Name() + "/" + subitem.Name())
							fmt.Println("Removed " + item.Name() + "/" + subitem.Name())
						}

					}
				}
			}
		}
	}
}
