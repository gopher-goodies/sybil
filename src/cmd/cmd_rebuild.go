package sybil_cmd

import "flag"
import "log"

import sybil "github.com/logv/sybil/src/lib"

// appends records to our record input queue
// every now and then, we should pack the input queue into a column, though
func RunRebuildCmdLine() {
	REPLACE_INFO := flag.Bool("replace", false, "Replace broken info.db if it exists")
	FORCE_UPDATE := flag.Bool("force", false, "Force re-calculation of info.db, even if it exists")
	flag.Parse()

	if *sybil.FLAGS.TABLE == "" {
		flag.PrintDefaults()
		return
	}

	if *sybil.FLAGS.PROFILE {
		profile := sybil.RUN_PROFILER()
		defer profile.Start().Stop()
	}

	t := sybil.GetTable(*sybil.FLAGS.TABLE)

	loaded := t.LoadTableInfo() && *FORCE_UPDATE == false
	if loaded {
		log.Println("TABLE INFO ALREADY EXISTS, NOTHING TO REBUILD!")
		return
	}

	t.DeduceTableInfoFromBlocks()

	// TODO: prompt to see if this table info looks good and then write it to
	// original info.db
	if *REPLACE_INFO == true {
		log.Println("REPLACING info.db WITH DATA COMPUTED ABOVE")
		lock := sybil.Lock{Table: t, Name: "info"}
		lock.ForceDeleteFile()
		t.SaveTableInfo("info")
	} else {
		log.Println("SAVING TO temp_info.db")
		t.SaveTableInfo("temp_info")
	}
}
