package torrent

//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrent --package-comment "DO NOT EDIT" --typename Info --o info.go data/info.json
//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrent --package-comment "DO NOT EDIT" --typename Properties --o properties.go data/properties.json
//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrent --package-comment "DO NOT EDIT" --typename Trackers --o trackers.go data/trackers.json
//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrent --package-comment "DO NOT EDIT" --typename File --o file.go data/file.json

type Files []File
