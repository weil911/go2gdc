package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"gdc"
	"github.com/json-iterator/go"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"tool"
)

//ReadDownloadFile reads the downloaded file.
func ReadDownloadFile(s0FilePath string) (m1FilepathB1 map[string][]byte) {
	m1FilepathB1 = make(map[string][]byte)
	if strings.HasSuffix(s0FilePath, ".tar.gz") {
		m1FilepathB1 = tool.ReadFileTarGz(s0FilePath)
		gdc.Md5sumByteCheck(m1FilepathB1)
	} else {
		m1FilepathB1 = tool.ReadFile(s0FilePath) // only 1 file, neither MANIFEST.txt, nor tar.gz, then no md5sum checking
	}
	return
}

//ReadDownloadByte reads the downloaded []byte.
func ReadDownloadByte(b1Body []byte) (m1FilepathB1 map[string][]byte) {
	m1FilepathB1 = make(map[string][]byte)
	readerGz, eGz := gzip.NewReader(bytes.NewReader(b1Body))
	if eGz != nil { // only 1 file
		m1FilepathB1["0"] = b1Body
	} else { // can be 1 .gz file or tar.gz
		readerTar := tar.NewReader(readerGz)
		_, eTar := readerTar.Next()
		if eTar != nil { // not tar.gz
			m1FilepathB1["0"] = b1Body
		} else {
			m1FilepathB1 = tool.ReadByteTarGz(b1Body)
			gdc.Md5sumByteCheck(m1FilepathB1)
		}
		defer readerGz.Close() // note: can not close when eGz != nil
	}
	return
}

//Usage shows usage of command line.
func Usage() {
	s1Cmd := []string{
		"go2gdc from={fromStatus}:{fromPath} to={toStatus}:{toPath}",
		"go2gdc from=filter:./path/to/filter.txt to=answer:./path/to/answer--",
		"go2gdc from=filter:./path/to/answer-.txt to=downloaded:./path/to/downloaded--",
		"go2gdc from=downloaded:./path/to/downloaded-.tar.gz to=integrated:./path/to/integrated--",
		"go2gdc from=filter:./path/to/filter.txt to=integrated:./path/to/integrated--",
	}
	for i, v := range s1Cmd {
		if i == 0 {
			fmt.Printf("Usage:\n%s\n\n", v)
		} else {
			fmt.Printf("example %d:\n%s\n\n", i, v)
		}
	}
	return
}

//Main is main function of this package.
func Main() {
	s0Sep := "--"
	s1Argument := os.Args
	if len(s1Argument) != 3 ||
		!strings.HasSuffix(s1Argument[0], "go2gdc") ||
		!strings.HasPrefix(s1Argument[1], "from=") ||
		!strings.HasPrefix(s1Argument[2], "to=") {
		fmt.Println("Error: wrong command")
		Usage()
		log.Fatalln()
	}
	s1From := strings.Split(strings.TrimPrefix(s1Argument[1], "from="), ":")
	s1To := strings.Split(strings.TrimPrefix(s1Argument[2], "to="), ":")
	if len(s1From) != 2 ||
		len(s1To) != 2 ||
		s1From[1] == "" ||
		s1To[1] == "" {
		fmt.Println("Error: missing {status} or {path}")
		Usage()
		log.Fatalln()
	}
	fmt.Println("From:", s1From)
	fmt.Println("To:", s1To)
	switch strings.Join([]string{s1From[0], s1To[0]}, ">") {
	case "filter>answer":
	case "filter>downloaded":
	case "filter>integrated":
	case "answer>downloaded":
	case "answer>integrated":
	case "downloaded>integrated":
	default:
		fmt.Println("Error: wrong \"{status}\" used")
		fmt.Println("1. The \"from={statusFrom}\" shoud be one of \"from=filter\", \"from=answer\" and \"from=downloaded\".")
		fmt.Println("2. The \"to={statusTo}\" shoud be one of \"to=answer\", \"to=downloaded\" and \"to=integrated\".")
		fmt.Println("3. The status in \"from={statusFrom}\" should be earlier than the status in \"to={statusTo}\".")
		log.Fatalln()
	}

	m3OmicstypeSampletypeFilter := make(map[string]map[string]map[string][]string)   // map[s0Omicstype]map[s0Sampletype]m1Filter
	m2OmicstypeSampletypeAnswer := make(map[string]map[string][]byte)                // map[s0Omicstype]map[s0Sampletype]b1Answer
	m3OmicstypeSampletypeFilepathB1 := make(map[string]map[string]map[string][]byte) // map[s0Omicstype]map[s0Sampletype]m1FilepathB1

	switch s1From[0] {
	case "filter":
		goto from_filter
	case "answer":
		s0PathAnswer := s1From[1]
		b1Answer := tool.ReadFile(s0PathAnswer)[s0PathAnswer]
		tool.CheckJson(b1Answer)
		b1CaseJson, b1CaseTsv := gdc.FilterAnswerToCase(b1Answer)
		s0FilePathCaseJson := strings.TrimSuffix(s0PathAnswer, ".json") + s0Sep + "Case.json"
		tool.SaveFile(s0FilePathCaseJson, b1CaseJson)
		s0FilePathCaseTsv := strings.TrimSuffix(s0PathAnswer, ".json") + s0Sep + "Case.tsv"
		tool.SaveFile(s0FilePathCaseTsv, b1CaseTsv)
		s0Omicstype := gdc.FilterAnswerToOmicstype(b1Answer)
		s0Sampletype := gdc.FilterAnswerToSampletype(b1Answer)
		if tool.FindS1([]string{"cnv_gene", "somatic_mutation"}, s0Omicstype) { // no Sampletype detail: cnv_gene, somatic_mutation
			s0Sampletype = "all"
		}
		m1SampletypeAnswer := make(map[string][]byte) // map[s0Omicstype]b1Answer
		m1SampletypeAnswer[s0Sampletype] = b1Answer
		m2OmicstypeSampletypeAnswer[s0Omicstype] = m1SampletypeAnswer
		goto from_answer
	case "downloaded":
		s0PathAnswer := ""
		if strings.HasSuffix(s1From[1], ".tar.gz") {
			s0PathAnswer = strings.TrimSuffix(s1From[1], ".tar.gz") + ".json"
		} else {
			s0PathAnswer = path.Dir(path.Dir(s1From[1])) + ".json" // for only 1 downloaded file
		}
		b1Answer := tool.ReadFile(s0PathAnswer)[s0PathAnswer]
		tool.CheckJson(b1Answer)
		s0Omicstype := gdc.FilterAnswerToOmicstype(b1Answer)
		s0Sampletype := gdc.FilterAnswerToSampletype(b1Answer)
		if tool.FindS1([]string{"cnv_gene", "somatic_mutation"}, s0Omicstype) { // no Sampletype detail: cnv_gene, somatic_mutation
			s0Sampletype = "all"
		}
		m1SampletypeAnswer := make(map[string][]byte) // map[s0Omicstype]b1Answer
		m1SampletypeAnswer[s0Sampletype] = b1Answer
		m2OmicstypeSampletypeAnswer[s0Omicstype] = m1SampletypeAnswer
		m1FilepathB1 := ReadDownloadFile(s1From[1])
		m2SampletypeFilepathB1 := make(map[string]map[string][]byte) // map[s0Omicstype]m1FilepathB1
		m2SampletypeFilepathB1[s0Sampletype] = m1FilepathB1
		m3OmicstypeSampletypeFilepathB1[s0Omicstype] = m2SampletypeFilepathB1
		goto from_downloaded
	}

from_filter:
	m3OmicstypeSampletypeFilter = gdc.FilterFileToM3S1(s1From[1])
	//fmt.Println("m3OmicstypeSampletypeFilter:", m3OmicstypeSampletypeFilter)
	for s0Omicstype, m2SampletypeFilter := range m3OmicstypeSampletypeFilter {
		m1SampletypeAnswer := make(map[string][]byte) // map[s0Omicstype]b1Answer
		for s0Sampletype, m1Filter := range m2SampletypeFilter {
			//fmt.Println("m2SampletypeFilter go2gdc: ", m2SampletypeFilter, "\n")
			//fmt.Println("m1Filter", m1Filter, "\n")
			//fmt.Println("m1Filter[\"cases.samples.sample_type_id\"]", m1Filter["cases.samples.sample_type_id"], "\n")
			b1Answer := gdc.FilterM1S1ToB1Answer(m1Filter)
			m1SampletypeAnswer[s0Sampletype] = b1Answer
			_, _, _, m1BarcodeProject := gdc.FilterAnswerInfo(b1Answer)
			s1Project := make([]string, 0)
			s1Barcode := make([]string, 0)
			for s0Barcode, s0Project := range m1BarcodeProject {
				if !tool.FindS1(s1Barcode, s0Barcode) {
					s1Barcode = append(s1Barcode, s0Barcode)
				}
				if !tool.FindS1(s1Project, s0Project) {
					s1Project = append(s1Project, s0Project)
				}
			}
			sort.Strings(s1Project)
			sort.Strings(s1Barcode)
			b1CaseJson, b1CaseTsv := gdc.FilterCase(s1Barcode)
			s0FilePathBase := s1To[1] + "Project_" + strings.Join(s1Project, "_") + s0Sep + "Omicstype_" + s0Omicstype + s0Sep + "Sampletype_" + s0Sampletype
			s0FilePath := s0FilePathBase + ".json"
			tool.SaveFile(s0FilePath, b1Answer)
			s0FilePathCaseJson := s0FilePathBase + s0Sep + "Case.json"
			tool.SaveFile(s0FilePathCaseJson, b1CaseJson)
			s0FilePathCaseTsv := s0FilePathBase + s0Sep + "Case.tsv"
			tool.SaveFile(s0FilePathCaseTsv, b1CaseTsv)
		}
		m2OmicstypeSampletypeAnswer[s0Omicstype] = m1SampletypeAnswer
	}
	if s1To[0] == "answer" {
		goto statusTo
	}

from_answer:
	for s0Omicstype, m1SampletypeAnswer := range m2OmicstypeSampletypeAnswer {
		m2SampletypeFilepathB1 := make(map[string]map[string][]byte) // map[s0Omicstype]m1FilepathB1
		for s0Sampletype, b1Answer := range m1SampletypeAnswer {
			s1Fileid, m1FilepathBarcode, _, m1BarcodeProject := gdc.FilterAnswerInfo(b1Answer)
			resp := gdc.Download(s1Fileid)
			b1Body := tool.ResponseToB1(resp)
			s1Project := make([]string, 0)
			for _, s0Project := range m1BarcodeProject {
				if !tool.FindS1(s1Project, s0Project) {
					s1Project = append(s1Project, s0Project)
				}
			}
			sort.Strings(s1Project)
			s0FilePathBase := s1To[1] + "Project_" + strings.Join(s1Project, "_") + s0Sep + "Omicstype_" + s0Omicstype + s0Sep + "Sampletype_" + s0Sampletype
			s0FilePath := s0FilePathBase + ".tar.gz"
			if len(m1FilepathBarcode) == 1 { // note: if only 1 file, neither MANIFEST.txt, nor tar.gz
				for k, _ := range m1FilepathBarcode {
					s0FilePath = s0FilePathBase + "/" + k
					break
				}
			}
			tool.SaveFile(s0FilePath, b1Body)
			m1FilepathB1 := ReadDownloadByte(b1Body)
			m2SampletypeFilepathB1[s0Sampletype] = m1FilepathB1
		}
		m3OmicstypeSampletypeFilepathB1[s0Omicstype] = m2SampletypeFilepathB1
	}
	if s1To[0] == "downloaded" {
		goto statusTo
	}

from_downloaded:
	for s0Omicstype, m2SampletypeFilepathB1 := range m3OmicstypeSampletypeFilepathB1 {
		//fmt.Println("s0Omicstype", s0Omicstype, "\n")
		for s0Sampletype, m1FilepathB1 := range m2SampletypeFilepathB1 {
			//fmt.Println("s0Sampletype", s0Sampletype, "\n")
			b1Answer := m2OmicstypeSampletypeAnswer[s0Omicstype][s0Sampletype]
			s1Fileid, m1FilepathBarcode, m1AliquotBarcode, m1BarcodeProject := gdc.FilterAnswerInfo(b1Answer)
			s1Project := make([]string, 0)
			for _, s0Project := range m1BarcodeProject {
				if !tool.FindS1(s1Project, s0Project) {
					s1Project = append(s1Project, s0Project)
				}
			}
			sort.Strings(s1Project)
			m1ColB1 := make(map[string][]byte)
			if s0Omicstype == "somatic_mutation" {
				m1ColB1 = gdc.TransM1B1Somatic(m1FilepathB1, s0Omicstype)
			} else {
				if len(m1FilepathB1) == 1 {
					for k, v := range m1FilepathB1 {
						delete(m1FilepathB1, k) // note: if being last line in cycle, all keys will be deleted, seems k being trailed
						m1FilepathB1[s1Fileid[0]] = v
					}
					for k, v := range m1FilepathBarcode {
						delete(m1FilepathBarcode, k) // note: if being last line in cycle, all keys will be deleted, seems k being trailed
						m1FilepathBarcode[strings.Split(k, "/")[0]] = v
					}
				}
				m1BarcodeB1 := gdc.TransM1B1KeyMap(m1FilepathB1, m1FilepathBarcode)
				m3FileColRowCell := gdc.TransM1B1ToM3S0(m1BarcodeB1, s0Omicstype)
				if s0Omicstype == "cnv_gene" {
					m3FileColRowCell2nd := make(map[string]map[string]map[string]string)
					for k1, v1 := range m3FileColRowCell {
						m2ColRowCell2nd := make(map[string]map[string]string)
						for k2, v2 := range v1 { // cnv_gene colname
							s0Barcode := m1AliquotBarcode[k2]
							m2ColRowCell2nd[m1BarcodeProject[s0Barcode]+"___"+s0Barcode] = v2
						}
						m3FileColRowCell2nd[k1] = m2ColRowCell2nd
					}
					m3FileColRowCell = m3FileColRowCell2nd
				}
				m1ColB1 = gdc.TransM3S0ToM1B1(m3FileColRowCell, s0Omicstype)
			}
			for k, v := range m1ColB1 {
				s0FilePathBase := s1To[1] + "Project_" + strings.Join(s1Project, "_") + s0Sep + "Omicstype_" + s0Omicstype + s0Sep + "Sampletype_" + s0Sampletype
				s0FilePath := s0FilePathBase + s0Sep + k + ".tsv"
				if tool.FindS1([]string{"cnv_gene", "somatic_mutation"}, s0Omicstype) {
					s0FilePath = s0FilePathBase + ".tsv"
				}
				tool.SaveFile(s0FilePath, v)
			}
		}
	}

	if s1To[0] == "integrated" {
		goto statusTo
	}

statusTo:
	return
}

//FileCount counts the files in every pair of Project-Omicstype.
func FileCount() {
	s0SaveDir := "~/test/gdc"
	s0SaveDir = path.Clean(s0SaveDir) + "/" // the dir ends with "/"
	tool.MakeDir(s0SaveDir)
	//gdc.EndpointSummary(s0SaveDir)
	s1Project := []string{
		//"BEATAML1.0-COHORT",
		//"BEATAML1.0-CRENOLANIB",
		//"CGCI-BLGSP",
		//"CPTAC-3",
		//"CTSP-DLBCL1",
		//"FM-AD",
		//"HCMI-CMDC",
		//"MMRF-COMMPASS",
		//"NCICCR-DLBCL",
		//"ORGANOID-PANCREATIC",
		"TARGET-ALL-P1",
		"TARGET-ALL-P2",
		"TARGET-ALL-P3",
		"TARGET-AML",
		//"TARGET-CCSK",
		//"TARGET-NBL",
		//"TARGET-OS",
		//"TARGET-RT",
		//"TARGET-WT",
		"TCGA-ACC",
		"TCGA-BLCA",
		//"TCGA-BRCA",
		//"TCGA-CESC",
		//"TCGA-CHOL",
		//"TCGA-COAD",
		//"TCGA-DLBC",
		//"TCGA-ESCA",
		//"TCGA-GBM",
		//"TCGA-HNSC",
		//"TCGA-KICH",
		//"TCGA-KIRC",
		//"TCGA-KIRP",
		//"TCGA-LAML",
		//"TCGA-LGG",
		//"TCGA-LIHC",
		//"TCGA-LUAD",
		//"TCGA-LUSC",
		//"TCGA-MESO",
		//"TCGA-OV",
		//"TCGA-PAAD",
		//"TCGA-PCPG",
		//"TCGA-PRAD",
		//"TCGA-READ",
		//"TCGA-SARC",
		//"TCGA-SKCM",
		//"TCGA-STAD",
		//"TCGA-TGCT",
		//"TCGA-THCA",
		//"TCGA-THYM",
		//"TCGA-UCEC",
		//"TCGA-UCS",
		//"TCGA-UVM",
	}
	s1Omicstype := []string{
		//"cnv_segment_somatic_only",
		//"cnv_segment_somatic_and_germline",
		//"cnv_gene",
		//"gene_htseq_fpkm_uq",
		//"gene_htseq_fpkm",
		//"gene_htseq_counts",
		//"gene_star_counts",
		//"methy_27",
		//"methy_450",
		"mir",
		"mir_isoform",
		//"somatic_mutation",
	}
	m2Count := make(map[string]map[string]string)
	for _, s0Omicstype := range s1Omicstype {
		m1Filter := gdc.FilterByOmicstype(s0Omicstype)
		m1Filter["access"] = []string{"open"}
		m1Count := make(map[string]string)
		for _, s0Project := range s1Project {
			m1Filter["cases.project.project_id"] = []string{s0Project}
			b1Hit := gdc.EndpointRecordsAllFiltered("files", m1Filter, s0SaveDir+"/"+s0Omicstype+"-"+s0Project)
			m1Count[s0Project] = jsoniter.Get(b1Hit, "count").ToString()
			//fmt.Println("m1Count:", m1Count)
			//fmt.Println("b1Hit:", b1Hit)
		}
		m2Count[s0Omicstype] = m1Count
		//fmt.Println("m2Count:", m2Count)
	}
	b1Count := tool.TransS2ToB1(tool.TransM2S0ToS2(m2Count, "count", "21"), "\n", "\t")
	tool.SaveFile(s0SaveDir+"count"+".tsv", b1Count)
	return
}

func main() {
	Main()
	return
}

// BUG(): #2:
