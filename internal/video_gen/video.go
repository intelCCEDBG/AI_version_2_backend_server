package videogen

import (
	"os"
	"recorder/pkg/logger"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GenerateVideo(hour int, minute int, duration int, hostname string) {
	files, err := os.ReadDir("/home/media/video/" + hostname + "/")
	if err != nil {
		logger.Error("List video dir fail: " + err.Error())
	}
	currentTime := time.Now()
	today_date_string := currentTime.Format("2006-01-02")
	yesterday_date_string := currentTime.AddDate(0, 0, -1).Format("2006-01-02")
	// fmt.Println(today_date_string)
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	sort.Strings(filenames)
	f, err := os.OpenFile("/home/media/video/"+hostname+"/self-define.m3u8", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 644)
	if err != nil {
		logger.Error("open video file fail: " + err.Error())
		return
	}
	defer f.Close()
	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:10\n")
	var h, m string
	if hour < 10 {
		h = strconv.Itoa(hour)
		h = "0" + h
	} else {
		h = strconv.Itoa(hour)
	}
	if minute < 10 {
		m = strconv.Itoa(minute)
		m = "0" + m
	} else {
		m = strconv.Itoa(minute)
	}
	reqtime := hour*100 + minute
	ctime := time.Now().Hour()*100 + time.Now().Minute()
	var datestring string
	if reqtime > ctime {
		datestring = today_date_string
	} else {
		datestring = yesterday_date_string
	}
	// fmt.Println(datestring)
	for ii, filename := range filenames {
		if strings.Contains(filename, datestring+"_"+h+"-"+m) {
			for i := 0; i < duration/10; i++ {
				f.WriteString("#EXTINF:10.000000,\n")
				f.WriteString(filenames[ii+i] + "\n")
			}
			break
		}
	}
	f.WriteString("#EXT-X-ENDLIST")
}
