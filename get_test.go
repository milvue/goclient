package goclient

import (
	"fmt"
	"testing"

	"github.com/rronan/gonetdicom/dicomutil"
)

func Test_Get(t *testing.T) {
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		fmt.Println(inference_command)
		dcm_slice, msg, err := Get(API_URL, "1.2.840.113970.1.2.840.113970.6418804.20201101.1205635", inference_command, TOKEN)
		if err != nil {
			panic(err)
		}
		for _, dcm := range dcm_slice {
			study_instance_uid, sop_instance_uid, err := dicomutil.GetUIDs(dcm)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s,%s\n", study_instance_uid, sop_instance_uid)
		}
		print(msg)
	}
}

func Test_GetSignedUrl(t *testing.T) {
	for _, inference_command := range []string{"smarturgences", "smartxpert"} {
		dcm_slice, msg, err := GetSignedUrl(API_URL, "1.2.840.113970.1.2.840.113970.6418804.20201101.1205635", inference_command, TOKEN)
		if err != nil {
			panic(err)
		}
		for _, dcm := range dcm_slice {
			study_instance_uid, sop_instance_uid, err := dicomutil.GetUIDs(dcm)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s,%s\n", study_instance_uid, sop_instance_uid)
		}
		print(msg)
	}
}
