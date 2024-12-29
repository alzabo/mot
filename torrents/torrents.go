package torrents

//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrents --package-comment "DO NOT EDIT" --typename Info --o info.go data/info.json
//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrents --package-comment "DO NOT EDIT" --typename Properties --o properties.go data/properties.json
//go:generate gojsonstruct --abbreviations TMM,DL,ETA,NB --int-type int64 --package-name torrents --package-comment "DO NOT EDIT" --typename Trackers --o trackers.go data/trackers.json
