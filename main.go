package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/barasher/go-exiftool"
)

func main() {
	logFile, err := os.OpenFile(timeNow.Format("20060102150405")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(mw, "INFO: ", log.Flags()&^(log.Ldate|log.Ltime))
	defer logFile.Close()

	logger.Println("yaRenamer started!")

	checkEt(logger)
	workDir := getConfig(logger)
	dirFiles, forExifTool := walkingOnFilesystem(workDir, logger)
	if len(dirFiles)+len(forExifTool) == 0 {
		logger.Println("Nothin to do!\nBye :)")
		os.Exit(0)
	}
	if len(dirFiles) > 0 {
		mustCompile := regexp.MustCompile(`.*(\d{4})[\._:-]?(\d{2})[\._:-]?(\d{2})[\._:-]?\s?(\d{2})[\._:-]?(\d{2})[\._:-]?(\d{2}).*`)
		for _, item := range dirFiles {
			logger.SetPrefix(filepath.Base(item) + " ")
			nameSlice := mustCompile.FindStringSubmatch(filepath.Base(item))
			if len(nameSlice) == 0 {
				logger.Println("doing func:fsTimeStamp; when DateInName data corrupted")
				useFSTimeStamp(item, logger)
			}
			parsedFileYear, err := strconv.ParseInt(nameSlice[1], 10, 32)
			check(err)
			if parsedFileYear > int64(timeNow.Year()) || parsedFileYear < 1995 {
				logger.Println("Failed when parsed fileYear: ", parsedFileYear, "moved to exifRenamer func")
				forExifTool = append(forExifTool, item)
				continue
			}
			newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4] + nameSlice[5] + nameSlice[6]
			renamer(item, newName, logger)
			logger.Println("New name is a: " + newName + "of file:" + item)
		}
	}
	if len(forExifTool) > 0 && exiftoolExist {
		et, err := exiftool.NewExiftool()
		if err != nil {
			panic(fmt.Errorf("Error when intializing: %v", err))
		}
		defer et.Close()
		for _, item := range forExifTool {
			logger.SetPrefix(filepath.Base(item) + " ")
			exifData, err := getExif(et, item, logger)
			if err != nil { //if getExif data is failed
				logger.Println(err)
				logger.Println("func:fsTimeStamp; when exif data corrupted.")
				useFSTimeStamp(item, logger)
			} else {
				//etYear, err := strconv.ParseInt(stdLongYear, 10, 32) ////!!!!!!!!!!stdLo
				//check(err)
				if int64(exifData.Year()) > int64(timeNow.Year()) || int64(exifData.Year()) < 1995 {
					logger.Println("func:fsTimeStamp; when exif data falsified: ", stdLongYear)
					useFSTimeStamp(item, logger)
				} else {
					newName := exifData.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
					renamer(item, newName, logger)
					logger.Println("main:exifToolRename; newName: " + newName)
				}
			}
		}
	} else {
		if !exiftoolExist {
			logger.Println("SKIPPED: ", len(forExifTool), " files in ExifTool processing")
		}
	}
}
