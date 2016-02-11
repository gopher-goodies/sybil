package pcs

import "path"
import "log"
import "fmt"
import "os"
import "strings"

// there exists two dirs for ingesting and digesting:
// ingest/
// digest/

// to ingest, make a new tmp file inside ingest/ (or append to an existing one)
// to digest, move that file into digest/ and begin digesting it
// POTENTIAL RACE CONDITION ON INGEST MODIFYING AN EXISTING FILE!

// Go through newRecords list and save all the new records out to a row store
func (t *Table) IngestRecords(blockname string) {
	log.Println("KEY TABLE", t.KeyTable)

	t.AppendRecordsToLog(t.newRecords[:], blockname)
	t.newRecords = make(RecordList, 0)
	t.SaveTableInfo("info")
}

// Go through rowstore and save records out to column store
func (t *Table) DigestRecords(digest string) {
	// TODO: REFUSE TO DIGEST IF THE DIGEST AREA ALREADY EXISTS
	dirname := path.Join(*f_DIR, t.Name, "ingest")

	file, err := os.Open(dirname)
	if err != nil {
		log.Println("Can't open the ingestion dir", dirname)
		return
	}

	files, err := file.Readdir(0)
	digestdir := path.Join(*f_DIR, t.Name, "digest")
	digestname := path.Join(digestdir, fmt.Sprintf("%s.db", digest))
	os.MkdirAll(digestdir, 0777)
	handled := make([]string, 0)

	for _, filename := range files {

		if strings.HasPrefix(filename.Name(), digest) == false {
			continue
		}
		if strings.HasSuffix(filename.Name(), ".db") == false {
			continue
		}

		log.Println("Moving", filename, "TO", digestname, "FOR DIGESTION")
		fullname := fmt.Sprintf("%s/%s", dirname, filename.Name())

		err := os.Rename(fullname, digestname)
		if err != nil {
			log.Println("NO INGEST LOG, ERR:", err)
			return
		}

		records := t.LoadRecordsFromLog(digestname)
		log.Println("LOADED", len(records), "FOR DIGESTION")

		if len(records) > 0 {
			t.newRecords = append(t.newRecords, records...)
		}

		handled = append(handled, digestname)

		if len(t.newRecords) > 1000 {
			t.SaveRecords()
			// Release the records that were in this block, too...
			t.ReleaseRecords()

			for _, d := range handled {
				os.Remove(d)
				log.Println("Removing", d)
			}

			handled = handled[:0]
			t.newRecords = t.newRecords[:0]
		}
	}

	if len(t.newRecords) > 0 {
		t.SaveRecords()
		// Release the records that were in this block, too...
		t.ReleaseRecords()
	}

}
