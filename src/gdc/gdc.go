package gdc

import (
	"encoding/csv"
	"fmt"
	"github.com/json-iterator/go"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"tool"
)

var (
	json     = jsoniter.ConfigCompatibleWithStandardLibrary
	s0UrlGdc = "https://api.gdc.cancer.gov"
)

//CheckWarnings checks warnings in json format response from GDC.
func CheckWarnings(b1Json []byte) {
	tool.CheckJson(b1Json)
	for _, v := range jsoniter.Get(b1Json).Keys() {
		if v == "warnings" {
			warnings := jsoniter.Get(b1Json, v).ToString()
			if warnings != "{}" {
				log.Fatal(warnings)
			}
			break
		}
	}
	return
}

//EndpointRecordsDefault gets default response from GDC API endpoint.
func EndpointRecordsDefault(s0Endpoint string) (b1Body []byte) {
	s0Url := s0UrlGdc + "/" + s0Endpoint
	resp, e := http.Get(s0Url)
	tool.CheckError(e)
	b1Body = tool.ResponseToB1(resp)
	tool.CheckJson(b1Body)
	CheckWarnings(b1Body)
	return
}

//EndpointSummary gets summary of endpoint information: default, _mapping.
func EndpointSummary(s0PathPrefix string) {
	for _, s0Endpoint := range []string{
		"status",
		"projects",
		"annotations",
		"cases",
		"files",
	} {
		fmt.Println(time.Now(), ": endpoint", s0Endpoint)
		b1Body := EndpointRecordsDefault(s0Endpoint)
		tool.SaveFile(s0PathPrefix+"-default-"+s0Endpoint+".json", b1Body)
		if s0Endpoint == "status" { //no "_mapping" for "status"
			continue
		}
		b1Body = EndpointFieldsMapping(s0Endpoint)
		tool.SaveFile(s0PathPrefix+"-mapping-"+s0Endpoint+".json", b1Body)
	}
	return
}

//EndpointRecords gets selected endpoint records.
func EndpointRecords(s0Endpoint, s0Format, s0Sort string, i0From, i0Size int, s1Field []string, m1Filter map[string][]string) (b1Body []byte) {
	m1Payload := Payload(s0Format, s0Sort, i0From, i0Size, s1Field, m1Filter)
	b1Payload, e := json.Marshal(m1Payload)
	tool.CheckError(e)
	s0Url := s0UrlGdc + "/" + s0Endpoint
	b1Body = tool.ResponseToB1(tool.Post(s0Url, b1Payload, "application/json"))
	if s0Format == "json" {
		tool.CheckJson(b1Body)
	}
	CheckWarnings(b1Body)
	return
}

//EndpointRecordsAllFiltered gets all records (with default fields) of a endpoint.
func EndpointRecordsAllFiltered(s0Endpoint string, m1Filter map[string][]string, s0PathPrefix string) (b1Hit []byte) {
	switch s0Endpoint {
	case "projects",
		"annotations",
		"cases",
		"files":
		s0Sort := strings.TrimRight(s0Endpoint, "s") + "_id"
		s0Sort = "" //sort only usable for tsv format?
		s1Field := EndpointFieldsSelected(s0Endpoint)
		b1Body := EndpointRecords(s0Endpoint, "json", "", 0, 10, nil, m1Filter)
		i0Total := jsoniter.Get(b1Body, "data", "pagination", "total").ToInt()
		i0Size := 20000
		s1Hit := make([]string, i0Total/i0Size+1)
		for iFrom := 0; iFrom < i0Total; iFrom += i0Size {
			//fmt.Println(time.Now(), ": ", "total=", i0Total, ":", iFrom, "-", iFrom+i0Size)
			b1Body = EndpointRecords(s0Endpoint, "json", s0Sort, iFrom, i0Size, s1Field, m1Filter)
			if s0PathPrefix != "" {
				tool.SaveFile(s0PathPrefix+"-records-"+s0Endpoint+"-size"+strconv.Itoa(i0Size)+"from"+strconv.Itoa(iFrom)+".json", b1Body)
			}
			s1Hit[iFrom/i0Size] = strings.Trim(jsoniter.Get(b1Body, "data", "hits").ToString(), "[]")
		}
		b1Hit = []byte("{" + `"data":[` + strings.Join(s1Hit, ",") + `],"count":` + strconv.Itoa(i0Total) + "}")
		tool.CheckJson(b1Hit)
		if s0PathPrefix != "" {
			tool.SaveFile(s0PathPrefix+"-records-"+s0Endpoint+".json", b1Hit)
		}
	default:
		log.Fatalf("wrong endpoint: %s", s0Endpoint)
	}
	return
}

//FilterByOmicstype returns filter by omics-type.
func FilterByOmicstype(s0Omicstype string) (m1Filter map[string][]string) {
	s0Omicstype = strings.ToLower(s0Omicstype)
	m1Filter = make(map[string][]string)
	if s0Omicstype == "cnv_segment_somatic_only" || s0Omicstype == "cnv_segment_somatic_and_germline" || s0Omicstype == "cnv_gene" { // all "open"
		m1Filter["data_category"] = []string{
			"Copy Number Variation",
		}
		m1Filter["experimental_strategy"] = []string{
			"Genotyping Array",
		}
		switch s0Omicstype {
		case "cnv_gene":
			m1Filter["data_type"] = []string{
				"Gene Level Copy Number Scores", //cases/file, "<TCGA-project>.focal_score_by_genes.txt", 19730x?, TCGA- only, count 33 (1/cancer)
			}
		case "cnv_segment_somatic_and_germline":
			m1Filter["data_type"] = []string{
				"Copy Number Segment", //1case/file, ".grch38.seg.v2.txt", ?x6, TCGA- only, somatic and germline
			}
		case "cnv_segment_somatic_only":
			m1Filter["data_type"] = []string{
				"Masked Copy Number Segment", //1case/file, ".nocnv_grch38.seg.v2.txt", ?x6, TCGA- only, somatic only
			}
		}
		m1Filter["type"] = []string{ // can be annotated
			"copy_number_segment",  //1case/file, ".grch38.seg.v2.txt" or ".nocnv_grch38.seg.v2.txt"
			"copy_number_estimate", //cases/file, "<TCGA-project>.focal_score_by_genes.txt"
		}
		m1Filter["platform"] = []string{
			"Affymetrix SNP 6.0",
		}
	} else if s0Omicstype == "gene_htseq_fpkm_uq" || s0Omicstype == "gene_htseq_fpkm" || s0Omicstype == "gene_htseq_counts" || s0Omicstype == "gene_star_counts" {
		m1Filter["data_category"] = []string{
			"Transcriptome Profiling",
		}
		m1Filter["data_type"] = []string{
			"Gene Expression Quantification",
			"Splice Junction Quantification", //controlled
		}
		m1Filter["type"] = []string{
			"gene_expression",
		}
		m1Filter["experimental_strategy"] = []string{
			"RNA-Seq",
		}
		switch s0Omicstype {
		case "gene_htseq_fpkm_uq":
			m1Filter["analysis.workflow_type"] = []string{
				"HTSeq - FPKM-UQ",
			}
		case "gene_htseq_fpkm":
			m1Filter["analysis.workflow_type"] = []string{
				"HTSeq - FPKM",
			}
		case "gene_htseq_counts":
			m1Filter["analysis.workflow_type"] = []string{
				"HTSeq - Counts",
			}
		case "gene_star_counts":
			m1Filter["analysis.workflow_type"] = []string{
				"STAR - Counts",
			}
		}
	} else if s0Omicstype == "methy_450" { // all "open"
		m1Filter["data_category"] = []string{
			"DNA Methylation",
		}
		m1Filter["data_type"] = []string{
			"Methylation Beta Value",
		}
		m1Filter["experimental_strategy"] = []string{
			"Methylation Array",
		}
		m1Filter["type"] = []string{
			"methylation_beta_value",
		}
		m1Filter["platform"] = []string{
			"Illumina Human Methylation 450",
		}
	} else if s0Omicstype == "methy_27" { // all "open"
		m1Filter["data_category"] = []string{
			"DNA Methylation",
		}
		m1Filter["data_type"] = []string{
			"Methylation Beta Value",
		}
		m1Filter["experimental_strategy"] = []string{
			"Methylation Array",
		}
		m1Filter["type"] = []string{
			"methylation_beta_value",
		}
		m1Filter["platform"] = []string{
			"Illumina Human Methylation 27",
		}
	} else if s0Omicstype == "mir" {
		m1Filter["data_category"] = []string{
			"Transcriptome Profiling",
		}
		m1Filter["data_type"] = []string{
			"miRNA Expression Quantification",
		}
		m1Filter["experimental_strategy"] = []string{
			"miRNA-Seq",
		}
		m1Filter["type"] = []string{
			"mirna_expression",
		}
	} else if s0Omicstype == "mir_isoform" {
		m1Filter["data_category"] = []string{
			"Transcriptome Profiling",
		}
		m1Filter["data_type"] = []string{
			"Isoform Expression Quantification",
		}
		m1Filter["experimental_strategy"] = []string{
			"miRNA-Seq",
		}
		m1Filter["type"] = []string{
			"mirna_expression",
		}
	} else if s0Omicstype == "somatic_mutation" {
		m1Filter["data_category"] = []string{
			"Simple Nucleotide Variation",
		}
		m1Filter["experimental_strategy"] = []string{
			"Targeted Sequencing", //controlled
			"WXS",                 //open
		}
		m1Filter["data_format"] = []string{
			"MAF", //open
			"VCF", //controlled
		}
		m1Filter["data_type"] = []string{
			"Aggregated Somatic Mutation", //controlled
			"Annotated Somatic Mutation",  //controlled
			"Masked Somatic Mutation",     //open
			"Raw Simple Somatic Mutation", //controlled
		}
		m1Filter["type"] = []string{
			"aggregated_somatic_mutation", //controlled
			"annotated_somatic_mutation",  //controlled
			"masked_somatic_mutation",     //open
			"simple_somatic_mutation",     //controlled
		}
	} else if s0Omicstype == "reads" { // all "controlled"
		m1Filter["data_category"] = []string{
			"Sequencing Reads",
		}
	}
	return
}

//FilterAnswerToSampletype detects sample-type from json format answer.
func FilterAnswerToSampletype(b1Answer []byte) (s0Sampletype string) {
	_, _, _, m1BarcodeProject := FilterAnswerInfo(b1Answer)
	s1Sampletype := make([]string, 0)
	for s0Barcode, s0Project := range m1BarcodeProject {
		if strings.HasPrefix(s0Project, "TCGA-") {
			s1Sampletype = append(s1Sampletype, strings.Join(strings.Split(s0Barcode, "")[13:15], ""))
		} else {
			s0Sampletype = "notTCGA"
		}
	}
	s1Sampletype2nd := tool.SetS1(s1Sampletype)
	sort.Strings(s1Sampletype2nd)
	s0Sampletype = strings.Join(s1Sampletype2nd, "_")
	return
}

//FilterAnswerToOmicstype detects omics-type from json format answer.
func FilterAnswerToOmicstype(b1Answer []byte) (s0Omicstype string) {
	_, m1FilepathBarcode, _, _ := FilterAnswerInfo(b1Answer)
	s1Omicstype := make([]string, 0)
	for k, _ := range m1FilepathBarcode {
		if strings.HasSuffix(k, ".focal_score_by_genes.txt") {
			s1Omicstype = append(s1Omicstype, "cnv_gene")
		} else if strings.HasSuffix(k, ".grch38.seg.v2.txt") {
			s1Omicstype = append(s1Omicstype, "cnv_segment_somatic_and_germline")
		} else if strings.HasSuffix(k, ".nocnv_grch38.seg.v2.txt") {
			s1Omicstype = append(s1Omicstype, "cnv_segment_somatic_only")
		} else if strings.HasSuffix(k, ".FPKM-UQ.txt.gz") {
			s1Omicstype = append(s1Omicstype, "gene_htseq_fpkm_uq")
		} else if strings.HasSuffix(k, ".FPKM.txt.gz") {
			s1Omicstype = append(s1Omicstype, "gene_htseq_fpkm")
		} else if strings.HasSuffix(k, ".htseq.counts.gz") ||
			strings.HasSuffix(k, ".htseq_counts.txt.gz") {
			s1Omicstype = append(s1Omicstype, "gene_htseq_counts")
		} else if strings.HasSuffix(k, ".rna_seq.star_gene_counts.tsv.gz") {
			s1Omicstype = append(s1Omicstype, "gene_star_counts")
		} else if strings.Contains(k, ".HumanMethylation450.") {
			s1Omicstype = append(s1Omicstype, "methy_450")
		} else if strings.Contains(k, ".HumanMethylation27.") {
			s1Omicstype = append(s1Omicstype, "methy_27")
		} else if strings.HasSuffix(k, ".mirbase21.mirnas.quantification.txt") ||
			strings.HasSuffix(k, ".mirnaseq.mirnas.quantification.txt") {
			s1Omicstype = append(s1Omicstype, "mir")
		} else if strings.HasSuffix(k, ".mirbase21.isoforms.quantification.txt") ||
			strings.HasSuffix(k, ".mirnaseq.isoforms.quantification.txt") {
			s1Omicstype = append(s1Omicstype, "mir_isoform")
		} else if strings.HasSuffix(k, ".DR-10.0.somatic.maf.gz") {
			s1Omicstype = append(s1Omicstype, "somatic_mutation")
		}
	}
	s1Omicstype2nd := tool.SetS1(s1Omicstype)
	if len(s1Omicstype2nd) > 1 {
		log.Fatalf("more than 1 omics_type detected: %s", s1Omicstype2nd)
	} else {
		s0Omicstype = s1Omicstype2nd[0]
	}
	return
}

//FilterFileToM1S1 transforms filter file into a map of slice.
func FilterFileToM1S1(s0PathFilter string) (m1s1Filter map[string][]string) {
	b1Filter := tool.ReadFile(s0PathFilter)[s0PathFilter]
	m1s1Filter = make(map[string][]string)
	for _, s1Row := range tool.TransB1ToS2(b1Filter, "\n", "=") {
		s0FilterField := ""
		s0FilterValue := ""
		if len(s1Row) > 2 {
			log.Fatalf("wrong filter file format (len): %s", s1Row)
		} else {
			s0FilterField = strings.ToLower(strings.TrimSpace(s1Row[0]))
			if len(s1Row) == 2 {
				s0FilterValue = strings.ToLower(strings.TrimSpace(strings.Trim(s1Row[1], ",")))
			}
		}
		s1FilterValue := make([]string, 0)
		for _, v := range strings.Split(s0FilterValue, ",") {
			vChecked := strings.ToLower(strings.TrimSpace(v))
			switch s0FilterField {
			case "omics_type",
				"sample_type_id", // more ? cases-records with "sample_type_id" field
				"sample_types_for_cases_intersection",
				"sample_types_for_separated_integration",
				"omics_types_for_cases_intersection",
				"keep_samples_from_same_case",
				"project_id":
				if s0FilterField == "project_id" {
					vChecked = strings.ToUpper(vChecked)
				}
				s1FilterValueRange := FilterValue(s0FilterField)
				if !tool.FindS1(s1FilterValueRange, vChecked) {
					log.Fatalf("error: the value of \"%s\" should be one of %s , not %s", s0FilterField, s1FilterValueRange, vChecked)
				}
			case "case_id":
				vChecked = strings.ToUpper(strings.TrimSpace(v))
			case "": // blank line
			default:
				log.Fatalf("wrong filter file format: %s", v)
			}
			s1FilterValue = append(s1FilterValue, vChecked)
		}
		m1s1Filter[s0FilterField] = s1FilterValue
	}
	delete(m1s1Filter, "") // blank line
	return
}

//FilterFileToM3S1 gets filter from file (e.g. ./filter.txt).
func FilterFileToM3S1(s0PathFilter string) (m3OmicstypeSampletypeFilter map[string]map[string]map[string][]string) {
	m1s1Filter := FilterFileToM1S1(s0PathFilter)
	bool0IntersectByOmicstype := false
	if tool.FindS1(m1s1Filter["omics_types_for_cases_intersection"], "t") ||
		tool.FindS1(m1s1Filter["omics_types_for_cases_intersection"], "true") {
		bool0IntersectByOmicstype = true
	}
	bool0Intersect := false
	if tool.FindS1(m1s1Filter["sample_types_for_cases_intersection"], "t") ||
		tool.FindS1(m1s1Filter["sample_types_for_cases_intersection"], "true") {
		bool0Intersect = true
	}
	bool0Seperated := false
	if tool.FindS1(m1s1Filter["sample_types_for_separated_integration"], "true") ||
		tool.FindS1(m1s1Filter["sample_types_for_separated_integration"], "t") {
		bool0Seperated = true
	}
	bool0SamplesFromSameCase := false
	if tool.FindS1(m1s1Filter["keep_samples_from_same_case"], "one") ||
		tool.FindS1(m1s1Filter["keep_samples_from_same_case"], "none") {
		bool0SamplesFromSameCase = true
	}
	for _, s0Project := range m1s1Filter["project_id"] {
		if !strings.HasPrefix(s0Project, "TCGA-") && (bool0Seperated || bool0Intersect || bool0IntersectByOmicstype || bool0SamplesFromSameCase) {
			log.Fatalf("the sample_type_id_for_* and omics_types_for_cases_intersection filters only work on TCGA projects, not %s.", s0Project)
		}
	}

	m3OmicstypeSampletypeFilter = make(map[string]map[string]map[string][]string)
	for _, s0Omicstype := range m1s1Filter["omics_type"] {
		m1Filter0th := FilterByOmicstype(s0Omicstype)
		m1Filter0th["access"] = []string{"open"}
		m1Filter0th["cases.project.project_id"] = m1s1Filter["project_id"]
		if !(len(m1s1Filter["case_id"]) == 1 && m1s1Filter["case_id"][0] == "") {
			m1Filter0th["cases.submitter_id"] = m1s1Filter["case_id"]
		}
		b1Answer := FilterM1S1ToB1Answer(m1Filter0th)
		_, _, _, m1BarcodeProject := FilterAnswerInfo(b1Answer)
		s1Barcode0th := tool.KeyM1S0(m1BarcodeProject)
		m1SampletypeBarcode := tool.SubS1ToM1S1(s1Barcode0th, 13, 15, 0, 28) // map[sample_type_id]barcode
		m1SampletypeSample := tool.SubS1ToM1S1(s1Barcode0th, 13, 15, 0, 16)  // map[sample_type_id]sample_id
		m1SampletypeCase := tool.SubS1ToM1S1(s1Barcode0th, 13, 15, 0, 12)    // map[sample_type_id]case_id
		if !(len(m1s1Filter["sample_type_id"]) == 1 && m1s1Filter["sample_type_id"][0] == "") {
			for k, _ := range m1SampletypeCase {
				if !tool.FindS1(m1s1Filter["sample_type_id"], k) {
					delete(m1SampletypeBarcode, k)
					delete(m1SampletypeSample, k)
					delete(m1SampletypeCase, k)
				}
			}
		}
		s1Sampletype := tool.KeyM1S1(m1SampletypeCase)
		m2SampletypeFilter := make(map[string]map[string][]string)
		for _, s0Sampletype := range s1Sampletype {
			m1FilterTmp := make(map[string][]string) // avoid to change m1Filter0th
			for k, v := range m1Filter0th {
				m1FilterTmp[k] = v
			}
			m1FilterTmp["cases.samples.portions.analytes.aliquots.submitter_id"] = m1SampletypeBarcode[s0Sampletype]
			m1FilterTmp["cases.samples.sample_type_id"] = []string{s0Sampletype}
			m1FilterTmp["cases.samples.submitter_id"] = m1SampletypeSample[s0Sampletype]
			m1FilterTmp["cases.submitter_id"] = m1SampletypeCase[s0Sampletype]
			m2SampletypeFilter[s0Sampletype] = m1FilterTmp // if "= m1Filter0th", the m1Filter0th will be changed
		}
		if bool0SamplesFromSameCase {
			m2SampletypeFilter = FilterBoolKeepSamplesFromSameCase(m2SampletypeFilter, m1s1Filter["keep_samples_from_same_case"])
		}
		if bool0Intersect { // only for TCGA
			m2SampletypeFilter = FilterBoolSampleTypesForCaseIntersection(m2SampletypeFilter, bool0Intersect)
		}
		m3OmicstypeSampletypeFilter[s0Omicstype] = m2SampletypeFilter
	}
	if bool0IntersectByOmicstype {
		m3OmicstypeSampletypeFilter = FilterBoolOmicsTypesForCaseIntersection(m3OmicstypeSampletypeFilter)
	}
	if !bool0Seperated {
		m3OmicstypeSampletypeFilter2nd := make(map[string]map[string]map[string][]string)
		for s0Omicstype, m2SampletypeFilter := range m3OmicstypeSampletypeFilter {
			m2SampletypeFilter2nd := FilterBoolSampleTypesForSeperatedIntegrationFalse(m2SampletypeFilter)
			m3OmicstypeSampletypeFilter2nd[s0Omicstype] = m2SampletypeFilter2nd
		}
		m3OmicstypeSampletypeFilter = m3OmicstypeSampletypeFilter2nd
	}
	m3OmicstypeSampletypeFilter3rd := make(map[string]map[string]map[string][]string) // avoid to change m3OmicstypeSampletypeFilter
	for k, v := range m3OmicstypeSampletypeFilter {
		m3OmicstypeSampletypeFilter3rd[k] = v
	}
	for s0Omicstype3rd, m2SampletypeFilter3rd := range m3OmicstypeSampletypeFilter3rd { // no Sampletype separated integration: cnv_gene, somatic_mutation
		if tool.FindS1([]string{"cnv_gene", "somatic_mutation"}, s0Omicstype3rd) {
			m2SampletypeFilter3rdTmp := make(map[string]map[string][]string)
			for _, v := range m2SampletypeFilter3rd {
				m2SampletypeFilter3rdTmp["all"] = v
			}
			m3OmicstypeSampletypeFilter[s0Omicstype3rd] = m2SampletypeFilter3rdTmp
		}
	}
	return
}

//FilterBoolOmicsTypesForCaseIntersection gets the intersection of cases with data from all omics types.
func FilterBoolOmicsTypesForCaseIntersection(m3OmicstypeSampletypeFilter map[string]map[string]map[string][]string) (m3OmicstypeSampletypeFilterTmp map[string]map[string]map[string][]string) {
	m2OmicstypeSampletypeCase := make(map[string]map[string][]string)
	for s0Omicstype, m2SampletypeFilter := range m3OmicstypeSampletypeFilter {
		m1SampletypeCase := make(map[string][]string)
		for s0Sampletype, m1Filter := range m2SampletypeFilter {
			m1SampletypeCase[s0Sampletype] = m1Filter["cases.submitter_id"]
		}
		m2OmicstypeSampletypeCase[s0Omicstype] = m1SampletypeCase
	}
	//fmt.Println("m2OmicstypeSampletypeCase:", m2OmicstypeSampletypeCase)
	m2SampletypeOmicstypeCase := tool.TransM2S1(m2OmicstypeSampletypeCase)
	m1SampletypeCase := make(map[string][]string)
	for s0Sampletype, m1OmicstypeCase := range m2SampletypeOmicstypeCase {
		m1SampletypeCase[s0Sampletype] = tool.SetInterM1S1ToS1(m1OmicstypeCase)
	}
	//fmt.Println("m1SampletypeCase:", m1SampletypeCase)
	m3OmicstypeSampletypeFilterTmp = make(map[string]map[string]map[string][]string)
	for s0Omicstype, m2SampletypeFilter := range m3OmicstypeSampletypeFilter {
		m2SampletypeFilter2nd := make(map[string]map[string][]string)
		for s0Sampletype, m1Filter := range m2SampletypeFilter {
			m1FilterTmp := make(map[string][]string) // avoid to change m1Filter
			for k, v := range m1Filter {
				if k != "cases.submitter_id" {
					m1FilterTmp[k] = v
				} else {
					m1FilterTmp[k] = m1SampletypeCase[s0Sampletype] // m1SampletypeCase[s0Sampletype] can be "", for none file
				}
			}
			m2SampletypeFilter2nd[s0Sampletype] = m1FilterTmp
		}
		m3OmicstypeSampletypeFilterTmp[s0Omicstype] = m2SampletypeFilter2nd
	}
	return
}

//FilterBoolSampleTypesForSeperatedIntegrationFalse merges the filters of different sample types.
func FilterBoolSampleTypesForSeperatedIntegrationFalse(m2SampletypeFilter map[string]map[string][]string) (m2SampletypeFilterTmp map[string]map[string][]string) {
	s1SampletypeTmp := make([]string, 0)
	m1FilterTmp := make(map[string][]string)
	for s0Sampletype, m1Filter := range m2SampletypeFilter {
		s1SampletypeTmp = append(s1SampletypeTmp, s0Sampletype) //s1SampletypeTmp for united "sample_type_id", which can be "unspecified" (blank or "")
		for k, _ := range m1Filter {
			m1FilterTmp[k] = make([]string, 0)
		}
	}
	for _, m1Filter := range m2SampletypeFilter {
		for k, v := range m1Filter {
			m1FilterTmp[k] = append(m1FilterTmp[k], v...)
		}
	}
	sort.Strings(s1SampletypeTmp)
	m2SampletypeFilterTmp = make(map[string]map[string][]string)
	m2SampletypeFilterTmp[strings.Join(s1SampletypeTmp, "_")] = m1FilterTmp // "sample_type_id" can be "unspecified" (blank or "")
	return
}

//FilterBoolSampleTypesForCaseIntersection get intersection of cases with data in all sample types.
func FilterBoolSampleTypesForCaseIntersection(m2SampletypeFilter map[string]map[string][]string, bool0Intersect bool) (m2SampletypeFilterTmp map[string]map[string][]string) {
	m1SampletypeCase := make(map[string][]string)
	for s0Sampletype, m1Filter := range m2SampletypeFilter {
		m1SampletypeCase[s0Sampletype] = m1Filter["cases.submitter_id"]
	}
	//fmt.Println("m1SampletypeCase :", m1SampletypeCase)
	s1Case := make([]string, 0)
	if bool0Intersect {
		s1Case = tool.SetInterM1S1ToS1(m1SampletypeCase)
	} else {
		s1Case = tool.SetUnionM1S1ToS1(m1SampletypeCase)
	}
	//fmt.Println("s1Case length:", len(s1Case))
	//fmt.Println("s1Case:", s1Case)
	m2SampletypeFilterTmp = make(map[string]map[string][]string)
	for s0Sampletype, m1Filter := range m2SampletypeFilter {
		m1FilterTmp := make(map[string][]string) // avoid to change m1Filter
		for k, v := range m1Filter {
			m1FilterTmp[k] = v
		}
		m1FilterTmp["cases.submitter_id"] = s1Case
		m2SampletypeFilterTmp[s0Sampletype] = m1FilterTmp
	}
	return
}

//FilterBoolKeepSamplesFromSameCase selects samples from same case.
func FilterBoolKeepSamplesFromSameCase(m2SampletypeFilter map[string]map[string][]string, s1SamplesFromSameCase []string) (m2SampletypeFilterTmp map[string]map[string][]string) {
	m2SampletypeFilterTmp = make(map[string]map[string][]string)
	for s0Sampletype, m1Filter := range m2SampletypeFilter {
		m1CaseBarcode := tool.SubS1ToM1S1(m1Filter["cases.samples.portions.analytes.aliquots.submitter_id"], 0, 12, 0, 28)
		m1CaseSample := tool.SubS1ToM1S1(m1Filter["cases.samples.submitter_id"], 0, 12, 0, 16)
		s1BarcodePicked := make([]string, 0)
		s1SamplePicked := make([]string, 0)
		for s0Case, s1Sample := range m1CaseSample {
			if len(s1Sample) == 1 {
				s1BarcodePicked = append(s1BarcodePicked, m1CaseBarcode[s0Case][0])
				s1SamplePicked = append(s1SamplePicked, s1Sample[0])
			} else {
				if tool.FindS1(s1SamplesFromSameCase, "one") {
					s1BarcodePicked = append(s1BarcodePicked, m1CaseBarcode[s0Case][0])
					s1SamplePicked = append(s1SamplePicked, s1Sample[0])
				} else if tool.FindS1(s1SamplesFromSameCase, "none") {
					continue
				}
			}
		}
		m1FilterTmp := make(map[string][]string) // avoid to change m1Filter
		for k, v := range m1Filter {
			m1FilterTmp[k] = v
		}
		m1FilterTmp["cases.samples.portions.analytes.aliquots.submitter_id"] = s1BarcodePicked
		m1FilterTmp["cases.samples.submitter_id"] = s1SamplePicked
		m1FilterTmp["cases.submitter_id"] = tool.SubS1(s1BarcodePicked, 0, 12)
		m2SampletypeFilterTmp[s0Sampletype] = m1FilterTmp
	}
	return
}

//FilterM1S1ToB1Answer gets answer body with filter.
func FilterM1S1ToB1Answer(m1Filter map[string][]string) (b1Body []byte) {
	s0Endpoint := "files"
	s1Field := []string{"file_id"}
	b1Body = EndpointRecords(s0Endpoint, "json", "", 0, 10, s1Field, m1Filter)
	i0Total := jsoniter.Get(b1Body, "data", "pagination", "total").ToInt()
	s1Field = EndpointFieldsSelected(s0Endpoint)
	b1Body = EndpointRecords(s0Endpoint, "json", "", 0, i0Total, s1Field, m1Filter)
	return
}

//FilterCaseTsv gets information from case-answer.
func FilterCaseTsv(b1Body []byte) (b1CaseTsv []byte) {
	i0Count := jsoniter.Get(b1Body, "count").ToInt()
	m2CaseInfo := make(map[string]map[string]string)
	for i := 0; i < i0Count; i++ {
		s0Case := jsoniter.Get(b1Body, "data", i, "submitter_id").ToString()
		s0CaseId := jsoniter.Get(b1Body, "data", i, "case_id").ToString()
		s0Project := jsoniter.Get(b1Body, "data", i, "project", "project_id").ToString()
		m1CaseInfo := make(map[string]string)
		m1CaseInfo["demographic-age_at_index"] = jsoniter.Get(b1Body, "data", i, "demographic", "age_at_index").ToString()
		m1CaseInfo["demographic-days_to_birth"] = jsoniter.Get(b1Body, "data", i, "demographic", "days_to_birth").ToString()
		m1CaseInfo["demographic-days_to_death"] = jsoniter.Get(b1Body, "data", i, "demographic", "days_to_death").ToString()
		m1CaseInfo["demographic-ethnicity"] = jsoniter.Get(b1Body, "data", i, "demographic", "ethnicity").ToString()
		m1CaseInfo["demographic-gender"] = jsoniter.Get(b1Body, "data", i, "demographic", "gender").ToString()
		m1CaseInfo["demographic-race"] = jsoniter.Get(b1Body, "data", i, "demographic", "race").ToString()
		m1CaseInfo["demographic-updated_datetime"] = jsoniter.Get(b1Body, "data", i, "demographic", "updated_datetime").ToString()
		m1CaseInfo["demographic-vital_status"] = jsoniter.Get(b1Body, "data", i, "demographic", "vital_status").ToString()
		m1CaseInfo["demographic-year_of_birth"] = jsoniter.Get(b1Body, "data", i, "demographic", "year_of_birth").ToString()
		m1CaseInfo["demographic-year_of_death"] = jsoniter.Get(b1Body, "data", i, "demographic", "year_of_death").ToString()
		m1CaseInfo["diagnoses-age_at_diagnosis"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "age_at_diagnosis").ToString()
		m1CaseInfo["diagnoses-ajcc_pathologic_m"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "ajcc_pathologic_m").ToString()
		m1CaseInfo["diagnoses-ajcc_pathologic_n"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "ajcc_pathologic_n").ToString()
		m1CaseInfo["diagnoses-ajcc_pathologic_stage"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "ajcc_pathologic_stage").ToString()
		m1CaseInfo["diagnoses-ajcc_pathologic_t"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "ajcc_pathologic_t").ToString()
		m1CaseInfo["diagnoses-ajcc_staging_system_edition"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "ajcc_staging_system_edition").ToString()
		m1CaseInfo["diagnoses-classification_of_tumor"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "classification_of_tumor").ToString()
		m1CaseInfo["diagnoses-days_to_diagnosis"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "days_to_diagnosis").ToString()
		m1CaseInfo["diagnoses-days_to_last_follow_up"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "days_to_last_follow_up").ToString()
		m1CaseInfo["diagnoses-days_to_last_known_disease_status"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "days_to_last_known_disease_status").ToString()
		m1CaseInfo["diagnoses-days_to_recurrence"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "days_to_recurrence").ToString()
		m1CaseInfo["diagnoses-diagnosis_id"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "diagnosis_id").ToString()
		m1CaseInfo["diagnoses-icd_10_code"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "icd_10_code").ToString()
		m1CaseInfo["diagnoses-last_known_disease_status"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "last_known_disease_status").ToString()
		m1CaseInfo["diagnoses-morphology"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "morphology").ToString()
		m1CaseInfo["diagnoses-primary_diagnosis"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "primary_diagnosis").ToString()
		m1CaseInfo["diagnoses-prior_malignancy"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "prior_malignancy").ToString()
		m1CaseInfo["diagnoses-prior_treatment"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "prior_treatment").ToString()
		m1CaseInfo["diagnoses-progression_or_recurrence"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "progression_or_recurrence").ToString()
		m1CaseInfo["diagnoses-site_of_resection_or_biopsy"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "site_of_resection_or_biopsy").ToString()
		m1CaseInfo["diagnoses-synchronous_malignancy"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "synchronous_malignancy").ToString()
		m1CaseInfo["diagnoses-tissue_or_organ_of_origin"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "tissue_or_organ_of_origin").ToString()
		m1CaseInfo["diagnoses-updated_datetime"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "updated_datetime").ToString()
		m1CaseInfo["diagnoses-year_of_diagnosis"] = jsoniter.Get(b1Body, "data", i, "diagnoses", 0, "year_of_diagnosis").ToString()
		m1CaseInfo["disease_type"] = jsoniter.Get(b1Body, "data", i, "disease_type").ToString()
		m1CaseInfo["exposures-alcohol_history"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "alcohol_history").ToString()
		m1CaseInfo["exposures-alcohol_intensity"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "alcohol_intensity").ToString()
		m1CaseInfo["exposures-bmi"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "bmi").ToString()
		m1CaseInfo["exposures-cigarettes_per_day"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "cigarettes_per_day").ToString()
		m1CaseInfo["exposures-exposure_id"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "exposure_id").ToString()
		m1CaseInfo["exposures-height"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "height").ToString()
		m1CaseInfo["exposures-pack_years_smoked"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "pack_years_smoked").ToString()
		m1CaseInfo["exposures-updated_datetime"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "updated_datetime").ToString()
		m1CaseInfo["exposures-weight"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "weight").ToString()
		m1CaseInfo["exposures-years_smoked"] = jsoniter.Get(b1Body, "data", i, "exposures", 0, "years_smoked").ToString()
		m1CaseInfo["primary_site"] = jsoniter.Get(b1Body, "data", i, "primary_site").ToString()
		m1CaseInfo["tissue_source_site_id"] = jsoniter.Get(b1Body, "data", i, "tissue_source_site", "tissue_source_site_id").ToString()
		m1CaseInfo["updated_datetime"] = jsoniter.Get(b1Body, "data", i, "updated_datetime").ToString()
		m1CaseInfo["project"] = s0Project
		m1CaseInfo["case_id"] = s0CaseId
		m1CaseInfo["case"] = s0Case
		m2CaseInfo[strings.Join([]string{s0Project, s0Case, s0CaseId}, "___")] = m1CaseInfo
	}
	s0IdColName := "Id"
	s0DimOrder := "12"
	b1CaseTsv = tool.TransS2ToB1(tool.TransM2S0ToS2(m2CaseInfo, s0IdColName, s0DimOrder), "\n", "\t")
	return
}

//FilterCase gets case-answer and case information from barcodes.
func FilterCase(s1Barcode []string) (b1CaseJson, b1CaseTsv []byte) {
	m1Filter := make(map[string][]string)
	m1Filter["cases.submitter_id"] = tool.SubS1(s1Barcode, 0, 12)
	b1CaseJson = EndpointRecordsAllFiltered("cases", m1Filter, "")
	b1CaseTsv = FilterCaseTsv(b1CaseJson)
	return
}

//FilterAnswerToCase extracts case information files from case-answer.
func FilterAnswerToCase(b1Answer []byte) (b1CaseJson, b1CaseTsv []byte) {
	_, _, _, m1BarcodeProject := FilterAnswerInfo(b1Answer)
	s1Barcode := make([]string, 0)
	for s0Barcode, _ := range m1BarcodeProject {
		s1Barcode = append(s1Barcode, s0Barcode)
	}
	sort.Strings(s1Barcode)
	b1CaseJson, b1CaseTsv = FilterCase(s1Barcode)
	return
}

//FilterAnswerInfo gets information from json format answer of filtered files.
func FilterAnswerInfo(b1Body []byte) (s1Fileid []string, m1FilepathBarcode, m1AliquotBarcode, m1BarcodeProject map[string]string) {
	i0Count := jsoniter.Get(b1Body, "data", "pagination", "count").ToInt()
	s1Fileid = make([]string, 0)
	s1Aliquot := make([]string, 0)
	m1FilepathBarcode = make(map[string]string)
	m1AliquotBarcode = make(map[string]string)
	m1BarcodeProject = make(map[string]string)
	for i := 0; i < i0Count; i++ {
		s0Project := jsoniter.Get(b1Body, "data", "hits", i, "cases", 0, "project", "project_id").ToString()
		s0Filename := jsoniter.Get(b1Body, "data", "hits", i, "file_name").ToString()
		s0Fileid := jsoniter.Get(b1Body, "data", "hits", i, "file_id").ToString()
		s1Fileid = append(s1Fileid, s0Fileid)
		s0FileidFilename := strings.Join([]string{s0Fileid, s0Filename}, "/") // like in tar.gz, even only 1 file downloaded
		s0Aliquot := ""
		s0Barcode := ""
		s1AliquotPerFile := make([]string, 0)
		for i2 := 0; ; i2++ {
			s0Aliquot = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", 0, "portions", 0, "analytes", 0, "aliquots", 0, "aliquot_id").ToString()
			s0Barcode = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", 0, "portions", 0, "analytes", 0, "aliquots", 0, "submitter_id").ToString()
			if s0Aliquot == "" && s0Barcode == "" {
				break
			}
			if !tool.FindS1(s1Aliquot, s0Aliquot) {
				s1Aliquot = append(s1Aliquot, s0Aliquot)
				s1AliquotPerFile = append(s1AliquotPerFile, s0Aliquot)
				m1AliquotBarcode[s0Aliquot] = s0Barcode
				m1BarcodeProject[s0Barcode] = s0Project
			}
			for i3 := 0; ; i3++ {
				s0Aliquot = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", 0, "analytes", 0, "aliquots", 0, "aliquot_id").ToString()
				s0Barcode = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", 0, "analytes", 0, "aliquots", 0, "submitter_id").ToString()
				if s0Aliquot == "" && s0Barcode == "" {
					break
				}
				if !tool.FindS1(s1Aliquot, s0Aliquot) {
					s1Aliquot = append(s1Aliquot, s0Aliquot)
					s1AliquotPerFile = append(s1AliquotPerFile, s0Aliquot)
					m1AliquotBarcode[s0Aliquot] = s0Barcode
					m1BarcodeProject[s0Barcode] = s0Project
				}
				for i4 := 0; ; i4++ {
					s0Aliquot = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", 0, "aliquots", 0, "aliquot_id").ToString()
					s0Barcode = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", 0, "aliquots", 0, "submitter_id").ToString()
					if s0Aliquot == "" && s0Barcode == "" {
						break
					}
					if !tool.FindS1(s1Aliquot, s0Aliquot) {
						s1Aliquot = append(s1Aliquot, s0Aliquot)
						s1AliquotPerFile = append(s1AliquotPerFile, s0Aliquot)
						m1AliquotBarcode[s0Aliquot] = s0Barcode
						m1BarcodeProject[s0Barcode] = s0Project
					}
					for i5 := 0; ; i5++ {
						s0Aliquot = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", i5, "aliquots", 0, "aliquot_id").ToString()
						s0Barcode = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", i5, "aliquots", 0, "submitter_id").ToString()
						if s0Aliquot == "" && s0Barcode == "" {
							break
						}
						if !tool.FindS1(s1Aliquot, s0Aliquot) {
							s1Aliquot = append(s1Aliquot, s0Aliquot)
							s1AliquotPerFile = append(s1AliquotPerFile, s0Aliquot)
							m1AliquotBarcode[s0Aliquot] = s0Barcode
							m1BarcodeProject[s0Barcode] = s0Project
						}
						for i6 := 0; ; i6++ {
							s0Aliquot = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", i5, "aliquots", i6, "aliquot_id").ToString()
							s0Barcode = jsoniter.Get(b1Body, "data", "hits", i, "cases", i2, "samples", i3, "portions", i4, "analytes", i5, "aliquots", i6, "submitter_id").ToString()
							if s0Aliquot == "" && s0Barcode == "" {
								break
							}
							if !tool.FindS1(s1Aliquot, s0Aliquot) {
								s1Aliquot = append(s1Aliquot, s0Aliquot)
								s1AliquotPerFile = append(s1AliquotPerFile, s0Aliquot)
								m1AliquotBarcode[s0Aliquot] = s0Barcode
								m1BarcodeProject[s0Barcode] = s0Project
							}
						}
					}
				}
			}
		}
		if len(s1AliquotPerFile) == 1 {
			m1FilepathBarcode[s0FileidFilename] = strings.Join([]string{s0Project, m1AliquotBarcode[s1Aliquot[i]]}, "___")
		} else {
			m1FilepathBarcode[s0FileidFilename] = strings.Join([]string{s0Project, strconv.Itoa(len(s1AliquotPerFile))}, "___")
		}
	}
	s1Barcode := make([]string, 0)
	for k, v := range m1AliquotBarcode {
		if tool.FindS1(s1Barcode, v) {
			log.Fatal("warning: repeat aliquot_id:barcode %s:%s", k, v)
		}
		s1Barcode = append(s1Barcode, v)
	}
	if len(m1BarcodeProject) != len(m1AliquotBarcode) {
		fmt.Printf("warning: length(unique barcode) != length(unique aliquot_id): %s!=%s", len(m1BarcodeProject), len(m1AliquotBarcode))
	}
	return
}

//PayloadFilterContent makes each content in filter.
func PayloadFilterContent(s0Field string, s1Value []string) (m1Content map[string]interface{}) {
	m1Content = map[string]interface{}{
		"op": "in",
		"content": map[string]interface{}{
			"field": s0Field,
			"value": s1Value,
		},
	}
	return
}

//PayloadFilterContents make all contents in filter.
func PayloadFilterContents(m1Filter map[string][]string) (m1sContent []map[string]interface{}) {
	for k, v := range m1Filter {
		m1sContent = append(m1sContent, PayloadFilterContent(k, v))
	}
	return
}

//Payload makes payload for post.
func Payload(s0Format, s0Sort string, i0From, i0Size int, s1Field []string, m1Filter map[string][]string) (m1Payload map[string]interface{}) {
	m1Payload = make(map[string]interface{})
	if s0Format != "" {
		m1Payload["format"] = s0Format
	}
	if s0Sort != "" {
		m1Payload["sort"] = s0Sort
	}
	if i0From != -1 {
		m1Payload["from"] = i0From
	}
	if i0Size != -1 {
		m1Payload["size"] = i0Size
	}
	if s1Field != nil {
		m1Payload["fields"] = strings.Join(s1Field, ",")
	}
	if m1Filter != nil {
		m1Payload["filters"] = map[string]interface{}{
			"op":      "and",
			"content": PayloadFilterContents(m1Filter),
		}
	}
	return
}

//Download downloads files with file_id: *.tar.gz if more than 1 file_id, or .../file_id/file if only 1 file_id.
func Download(s1Fileid []string) (resp *http.Response) {
	m1Payload := make(map[string][]string)
	m1Payload["ids"] = s1Fileid
	b1Payload, e := json.Marshal(m1Payload)
	tool.CheckError(e)
	s0Url := s0UrlGdc + "/data"
	resp = tool.Post(s0Url, b1Payload, "application/json")
	tool.CheckStatus(resp)
	return
}

//Manifest gets files info from "MANIFEST.txt".
func Manifest(s0FilePath string) (m1Manifest map[string][]string) {
	f, e := os.Open(s0FilePath)
	defer f.Close()
	tool.CheckError(e)
	readerCsv := csv.NewReader(f)
	readerCsv.Comment = '#'
	readerCsv.Comma = '\t'
	readerCsv.FieldsPerRecord = 5
	s2d, e := readerCsv.ReadAll()
	tool.CheckError(e)
	m1Manifest = make(map[string][]string)
	i0IndexId := -1
	i0IndexPath := -1
	i0IndexMd5 := -1
	for k1, v1 := range s2d {
		if k1 == 0 {
			for k2, v2 := range v1 {
				if v2 == "id" {
					i0IndexId = k2
					continue
				} else if v2 == "filename" {
					i0IndexPath = k2
					continue
				} else if v2 == "md5" {
					i0IndexMd5 = k2
					continue
				}
			}
		} else {
			if i0IndexId == -1 || i0IndexPath == -1 || i0IndexMd5 == -1 {
				log.Fatal("error: column id,filename,md5 not found in MANIFEST.TXT")
			}
			m1Manifest[v1[i0IndexId]] = []string{v1[i0IndexPath], v1[i0IndexMd5]}
		}
	}
	return
}

//Md5sumByteCheck checks md5 checksum in map[fileName]fileBody with "MANIFEST.txt" in it.
func Md5sumByteCheck(m1FileByte map[string][]byte) {
	s1SkipColName := []string{}
	s1SkipRowName := []string{}
	m2Manifest := tool.TransS2ToM2S0(tool.TransS2Skip(tool.TransB1ToS2(m1FileByte["MANIFEST.txt"], "\n", "\t"), s1SkipRowName, s1SkipColName), "21")
	for s0Filename, b1File := range m1FileByte {
		if s0Filename != "MANIFEST.txt" {
			if tool.Md5sumByte(b1File) != m2Manifest["md5"][path.Dir(s0Filename)] {
				log.Fatal("error: md5sum wrong")
			}
		}
	}
	return
}

//Md5sumFileCheck checks the untar dir with "MANIFEST.txt" and data files in sub-dir (md5 checksum) respectively.
func Md5sumFileCheck(s0UntarDir string) {
	s0UntarDir = path.Clean(s0UntarDir) + "/" // the dir ends with "/"
	m1Manifest := Manifest(s0UntarDir + "MANIFEST.txt")
	for _, v := range m1Manifest {
		if tool.Md5sumFile(s0UntarDir+v[0]) != v[1] {
			log.Fatal("error: md5sum wrong")
		}
	}
	return
}

//TransM1B1KeyMap changes map keys with key-map: map[fileId/fileName]fileData + map[fileId]barcode => map[barcode_fileId]fileData.
func TransM1B1KeyMap(m1B1 map[string][]byte, m1Key map[string]string) (m1B1Trans map[string][]byte) {
	s1Key := make([]string, 0, len(m1Key))
	m1Key2 := make(map[string]string)
	for k, v := range m1Key {
		k1 := strings.Split(k, "/")[0]
		s1Key = append(s1Key, k1)
		m1Key2[k1] = strings.Join([]string{v, strings.Split(k, "/")[0]}, "___") // ends with file_id
	}
	m1B1Trans = make(map[string][]byte)
	for k, v := range m1B1 {
		s1KeySplited := strings.Split(k, "/")
		s0Fileid := s1KeySplited[0]
		if k == "MANIFEST.txt" {
			continue
		}
		if !tool.FindS1(s1Key, s0Fileid) {
			log.Fatalf("the key-map has no key: %s", s0Fileid)
		}
		m1B1Trans[m1Key2[s0Fileid]] = v
	}
	return
}

//TransM1B1Somatic deals with somatic_mutation data.
func TransM1B1Somatic(m1FileByte map[string][]byte, s0Omicstype string) (m1B1 map[string][]byte) {
	switch s0Omicstype {
	case "cnv_segment_somatic_only",
		"cnv_segment_somatic_and_germline",
		"gene_htseq_fpkm_uq",
		"gene_htseq_fpkm",
		"gene_htseq_counts",
		"gene_star_counts",
		"methy_27",
		"methy_450",
		"mir",
		"mir_isoform",
		"cnv_gene":
		log.Fatalf("unsuitable omics type: %s", s0Omicstype)
	case "somatic_mutation":
		s1Head := make([]string, 0)   // head line
		s2Data := make([][]string, 0) // data lines
		for s0Filename, b1File := range m1FileByte {
			s1FileProjectPipe := make([]string, 3)
			if s0Filename == "MANIFEST.txt" {
				continue
			} else {
				s1Filename := strings.Split(s0Filename, ".")
				s1FileProjectPipe = []string{
					s1Filename[3], // file_id
					strings.Join(
						[]string{
							strings.Split(s1Filename[0], "/")[1],
							s1Filename[1]},
						"-"), // project_id
					s1Filename[2], // calling_pipeline
				}
			}
			s2Row := tool.TransB1ToS2(tool.ReadByteGz(b1File), "\n", "\t")
			for i, v := range s2Row { // no heading "#..." lines
				if strings.HasPrefix(v[0], "#") {
					continue
				} else {
					s2Row = s2Row[i:]
					break
				}
			}
			for i, v := range s2Row {
				if i == 0 { // head line
					if len(s1Head) == 0 { // first file
						s1Head = v
					} else if !tool.EqualS1(s1Head, v) { // check head lines from different pipelines
						log.Fatalf("different head line: %s != %s", s1Head, v)
					}
				} else { // data line
					s2Data = append(s2Data, append(s1FileProjectPipe, v...))
				}
			}
		}
		s2HeadData := make([][]string, len(s2Data)+1)
		s2HeadData[0] = append([]string{"file_id", "project_id", "calling_pipeline"}, s1Head...)
		for i, v := range s2Data {
			s2HeadData[i+1] = v
		}
		m1B1 = make(map[string][]byte)
		m1B1[s0Omicstype] = tool.TransS2ToB1(s2HeadData, "\n", "\t") // somatic_mutation filename
	default:
		log.Fatalf("wrong omics type: %s", s0Omicstype)
	}
	return
}

//TransM1B1ToM3S0 transforms map[string][]byte into map[string]map[string]map[string]string by omics-type.
func TransM1B1ToM3S0(m1FileByte map[string][]byte, s0Omicstype string) (m3FileColRowCell map[string]map[string]map[string]string) {
	m3FileColRowCell = make(map[string]map[string]map[string]string) // map[fileName]map[colName]map[rowName]cell
	for s0Filename, b1File := range m1FileByte {
		if s0Filename == "MANIFEST.txt" {
			continue
		}
		m2ColRowCell := make(map[string]map[string]string)
		s1SkipColName := []string{}
		s1SkipRowName := []string{}
		s2Row := make([][]string, 0)
		switch s0Omicstype {
		case "cnv_segment_somatic_only":
			s1SkipColName = []string{"GDC_Aliquot", "Num_Probes"}
			s2Row = tool.UniteS2(tool.TransB1ToS2(b1File, "\n", "\t"), []int{1, 2, 3}, "___")
		case "cnv_segment_somatic_and_germline":
			s1SkipColName = []string{"GDC_Aliquot", "Num_Probes"}
			s2Row = tool.UniteS2(tool.TransB1ToS2(b1File, "\n", "\t"), []int{1, 2, 3}, "___")
		case "cnv_gene":
			s1SkipColName = []string{"Gene ID", "Cytoband"}
			s2Row = tool.TransB1ToS2(b1File, "\n", "\t")
		case "gene_htseq_fpkm_uq":
			s1Row1st := []string{"Gene", "FPKMUQ"}
			s2Row = append([][]string{s1Row1st}, tool.TransB1ToS2(tool.ReadByteGz(b1File), "\n", "\t")...)
		case "gene_htseq_fpkm":
			s1Row1st := []string{"Gene", "FPKM"}
			s2Row = append([][]string{s1Row1st}, tool.TransB1ToS2(tool.ReadByteGz(b1File), "\n", "\t")...)
		case "gene_htseq_counts":
			s1Row1st := []string{"Gene", "ReadCount"}
			s2Row = append([][]string{s1Row1st}, tool.TransB1ToS2(tool.ReadByteGz(b1File), "\n", "\t")...)
			s1SkipRowName = []string{"__no_feature", "__ambiguous", "__too_low_aQual", "__not_aligned", "__alignment_not_unique"}
		case "gene_star_counts":
			s2Row = tool.TransB1ToS2(tool.ReadByteGz(b1File), "\n", "\t")
			s1SkipColName = []string{"stranded_first", "stranded_second"}
			s1SkipRowName = []string{"N_unmapped", "N_multimapping", "N_noFeature", "N_ambiguous"}
		case "methy_27", "methy_450":
			s2Row = tool.UniteS2(tool.TransB1ToS2(b1File, "\n", "\t"), []int{0, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "___")
		case "mir":
			s2Row = tool.TransB1ToS2(b1File, "\n", "\t")
		case "mir_isoform":
			s2Row = tool.UniteS2(tool.TransB1ToS2(b1File, "\n", "\t"), []int{0, 1, 5}, "___")
		case "somatic_mutation":
			log.Fatalf("unsuitable omics type: %s", s0Omicstype)
		}
		m2ColRowCell = tool.TransS2ToM2S0(tool.TransS2Skip(s2Row, s1SkipRowName, s1SkipColName), "21")
		m3FileColRowCell[strings.Split(s0Filename, "/")[0]] = m2ColRowCell
	}
	return
}

//TransM3S0ToM1B1 integrates files by omics.
func TransM3S0ToM1B1(m3FileColRowCell map[string]map[string]map[string]string, s0Omicstype string) (m1ColB1 map[string][]byte) {
	m3ColRowFileCell := make(map[string]map[string]map[string]string)
	if s0Omicstype != "cnv_gene" {
		m3ColRowFileCell = tool.TransM3S0Filled(m3FileColRowCell, "231")
	}
	if s0Omicstype == "cnv_gene" {
		m2ColRowCell := make(map[string]map[string]string)
		s1Project := make([]string, 0)
		s1Aliquot := make([]string, 0)
		for k1, v1 := range m3FileColRowCell {
			s0Fileid := strings.Split(k1, "___")[2]
			s1Project = append(s1Project, k1)
			for k2, v2 := range v1 {
				m2ColRowCell[strings.Join([]string{k2, s0Fileid}, "___")] = v2
				if tool.FindS1(s1Aliquot, k2) {
					log.Fatal("warning: repeat aliquot_id in column header %s", k2)
				} else {
					s1Aliquot = append(s1Aliquot, k2)
				}
			}
		}
		m3ColRowFileCell[strings.Join(s1Project, "-")] = m2ColRowCell
	}
	m1ColName := make(map[string]string) //select column and change column name
	s0IdColName := "Id"
	switch s0Omicstype {
	case "cnv_segment_somatic_only",
		"cnv_segment_somatic_and_germline":
		m1ColName["SegmentMean"] = "Segment_Mean"
	case "cnv_gene":
		for k, _ := range m3ColRowFileCell {
			m1ColName[k] = k // cnv_gene filename
		}
	case "gene_htseq_fpkm_uq":
		m1ColName["FPKMUQ"] = "FPKMUQ"
	case "gene_htseq_fpkm":
		m1ColName["FPKM"] = "FPKM"
	case "gene_htseq_counts":
		m1ColName["ReadCount"] = "ReadCount"
	case "gene_star_counts":
		m1ColName["ReadCount"] = "unstranded"
	case "methy_27",
		"methy_450":
		s0IdColName = "CGcluster#Chr#Start#End#GeneSymbol#GeneType#TranscriptID#PositionToTSS#CGIcoordinate#FeatureType"
		m1ColName["BetaValue"] = "Beta_value"
	case "mir",
		"mir_isoform":
		m1ColName["ReadCount"] = "read_count"
		m1ColName["ReadsPerMillionMirMapped"] = "reads_per_million_miRNA_mapped"
		m1ColName["CrossMapped"] = "cross-mapped"
	case "somatic_mutation":
		log.Fatalf("unsuitable omics type: %s", s0Omicstype)
	default:
		log.Fatalf("wrong omics type: %s", s0Omicstype)
	}
	m1ColB1 = make(map[string][]byte, len(m3ColRowFileCell))
	s0DimOrder := "12"
	switch s0Omicstype {
	case "cnv_segment_somatic_only",
		"cnv_segment_somatic_and_germline",
		"gene_htseq_fpkm_uq",
		"gene_htseq_fpkm",
		"gene_htseq_counts",
		"gene_star_counts",
		"methy_27",
		"methy_450",
		"mir",
		"mir_isoform":
		s0DimOrder = "12"
	case "cnv_gene":
		s0DimOrder = "21"
	case "somatic_mutation":
		log.Fatalf("unsuitable omics type: %s", s0Omicstype)
	default:
		log.Fatalf("wrong omics type: %s", s0Omicstype)
	}
	for k, v := range m1ColName {
		m1ColB1[k] = tool.TransS2ToB1(tool.TransM2S0ToS2(m3ColRowFileCell[v], s0IdColName, s0DimOrder), "\n", "\t")
	}
	return
}

// BUG(weil911): #1: .
