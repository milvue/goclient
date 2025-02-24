package milvuesdk

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/rronan/gonetdicom/dicomutil"
)

func Test_GetStatus(t *testing.T) {
	status_response, err := GetStatus(API_URL, StudyInstanceUID, TOKEN, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(status_response)
}

func Test_Get(t *testing.T) {
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		fmt.Println(inference_command)
		dcm_slice, err := Get(API_URL, StudyInstanceUID, inference_command, TOKEN, 0, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, dcm := range dcm_slice {
			study_instance_uid, series_instance_uid, sop_instance_uid, err := dicomutil.GetUIDs(dcm)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%s/%s/%s\n", study_instance_uid, series_instance_uid, sop_instance_uid)
		}
	}
}

func Test_GetToFile(t *testing.T) {
	OUTDIR := "../data/outputs"
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		fmt.Println(inference_command)
		dcm_path_slice, err := GetToFile(API_URL, StudyInstanceUID, inference_command, TOKEN, OUTDIR, 0, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, dcm_path := range dcm_path_slice {
			fmt.Println(dcm_path)
			study_instance_uid, series_instance_uid, sop_instance_uid, err := dicomutil.ParseFileUIDs(dcm_path)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%s/%s/%s\n", study_instance_uid, series_instance_uid, sop_instance_uid)
			err = os.Remove(dcm_path)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
func Test_GetSignedUrl(t *testing.T) {
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		fmt.Println(inference_command)
		dcm_slice, err := GetSignedUrl(API_URL, StudyInstanceUID, inference_command, TOKEN, 0, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, dcm := range dcm_slice {
			study_instance_uid, series_instance_uid, sop_instance_uid, err := dicomutil.GetUIDs(dcm)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%s/%s/%s\n", study_instance_uid, series_instance_uid, sop_instance_uid)
		}
	}
}

func Test_GetSignedUrlToFile(t *testing.T) {
	OUTDIR := "../data/outputs"
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		fmt.Println(inference_command)
		dcm_path_slice, err := GetSignedUrlToFile(API_URL, StudyInstanceUID, inference_command, TOKEN, OUTDIR, 0, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, dcm_path := range dcm_path_slice {
			study_instance_uid, series_instance_uid, sop_instance_uid, err := dicomutil.ParseFileUIDs(dcm_path)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%s/%s/%s\n", study_instance_uid, series_instance_uid, sop_instance_uid)
			err = os.Remove(dcm_path)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_GetSmarturgences(t *testing.T) {
	smarturgences_response, err := GetSmarturgences(API_URL, StudyInstanceUID, TOKEN, 0)
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.Marshal(&smarturgences_response)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(j))
}

func Test_GetSmartxpert(t *testing.T) {
	smartxpert_response, err := GetSmartxpert(API_URL, StudyInstanceUID, TOKEN, 0)
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.Marshal(&smartxpert_response)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(j))
}
