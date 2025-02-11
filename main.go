/*
Copyright © 2024 Ryan White

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/alzabo/mot/cmd"
)

var cpuprofile string = "cpu.prof"
var memprofile string = "mem.prof"

func main() {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	cmd.Execute()

	//f, err := os.Create(memprofile)

	//if err != nil {
	//	log.Fatal("could not create memory profile: ", err)
	//}

	//defer f.Close() // error handling omitted for example
	//runtime.GC()    // get up-to-date statistics

	//if err := pprof.WriteHeapProfile(f); err != nil {
	//	log.Fatal("could not write memory profile: ", err)
	//}
}
