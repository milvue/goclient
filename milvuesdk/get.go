package milvuesdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rronan/gonetdicom/dicomutil"
	"github.com/rronan/gonetdicom/dicomweb"
	"github.com/suyashkumar/dicom"
)

var ErrPredictionError = errors.New("PredictionError")
var ErrPredictionRunning = errors.New("PredictionRunning")
var ErrPredictionTimeout = errors.New("PredictionTimeout")
var ErrFormattingError = errors.New("FormattingError")

func isFormattingError(err error) bool {
	reqErr, isReqErr := err.(*dicomweb.RequestError)
	if !isReqErr {
		return false
	}
	if reqErr.Headers.Get("Content-Type") != "application/json" {
		return false
	}
	var resp_json map[string]string
	decode_err := json.Unmarshal(reqErr.Content, &resp_json)
	if decode_err != nil {
		return false
	}
	message, ok := resp_json["message"]
	return ok && message == "Error formatting study"
}

func WaitDone(api_url, study_instance_uid string, token string, interval int, total_wait_time int, timeout int) (GetStudyStatusResponseV3, error) {
	t1 := time.Now().Add(time.Duration(total_wait_time * 1e9))
	var status_response GetStudyStatusResponseV3
	for time.Now().Before(t1) {
		status_response, err := GetStatus(api_url, study_instance_uid, token, timeout)
		if err != nil {
			return GetStudyStatusResponseV3{}, err
		}
		if status_response.Status == "error" {
			return status_response, ErrPredictionError
		} else if status_response.Status == "done" {
			return status_response, nil
		}
		time.Sleep(time.Duration(interval * 1e9))
	}
	return status_response, ErrPredictionTimeout
}

func GetStatus(api_url, study_instance_uid string, token string, timeout int) (GetStudyStatusResponseV3, error) {
	url := fmt.Sprintf("%s/v3/studies/%s/status", api_url, study_instance_uid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GetStudyStatusResponseV3{}, err
	}
	req.Header.Set("x-goog-meta-owner", token)
	client := &http.Client{Timeout: time.Duration(timeout * 1e9)}
	resp, err := client.Do(req)
	if err != nil {
		return GetStudyStatusResponseV3{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GetStudyStatusResponseV3{}, &dicomweb.RequestError{StatusCode: resp.StatusCode, Err: errors.New(resp.Status)}
	}
	status_response := GetStudyStatusResponseV3{}
	json.NewDecoder(resp.Body).Decode(&status_response)
	return status_response, nil
}

func Get(api_url, study_instance_uid string, inference_command string, token string, timeout int, additional_params string) ([]*dicom.Dataset, error) {
	url := fmt.Sprintf(
		"%s/v3/studies/%s?inference_command=%s&signed_url=false%s",
		api_url,
		study_instance_uid,
		inference_command,
		additional_params,
	)
	headers := map[string]string{"x-goog-meta-owner": token, "Content-Type": "multipart/related; type=application/dicom"}
	dcm_slice, byte_slice, err := dicomweb.Wado(url, headers, timeout)
	if err != nil {
		if isFormattingError(err) {
			return []*dicom.Dataset{}, ErrFormattingError
		}
		return []*dicom.Dataset{}, err
	}
	if len(byte_slice) > 0 {
		var status_response GetStudyStatusResponseV3
		_ = json.Unmarshal(byte_slice, &status_response)
		if getenv("LOG_LEVEL", "INFO") == "DEBUG" {
			log.Printf("%.150s %s %s %s", string(byte_slice), status_response.StudyInstanceUID, status_response.Status, status_response.Version)
		}
		if status_response.Status == "running" {
			return []*dicom.Dataset{}, ErrPredictionRunning
		}
	}
	return dcm_slice, nil
}

func GetToFile(api_url, study_instance_uid string, inference_command string, token string, folder string, timeout int, additional_params string) ([]string, error) {
	url := fmt.Sprintf(
		"%s/v3/studies/%s?inference_command=%s&signed_url=false%s",
		api_url,
		study_instance_uid,
		inference_command,
		additional_params,
	)
	headers := map[string]string{"x-goog-meta-owner": token, "Content-Type": "multipart/related; type=application/dicom"}
	dcm_path_slice, byte_slice, err := dicomweb.WadoToFile(url, headers, folder, timeout)
	if err != nil {
		if isFormattingError(err) {
			return []string{}, ErrFormattingError
		}
		return []string{}, err
	}
	if len(byte_slice) > 0 {
		var status_response GetStudyStatusResponseV3
		_ = json.Unmarshal(byte_slice, &status_response)
		if status_response.Status == "running" {
			return []string{}, ErrPredictionRunning
		}
	}
	return dcm_path_slice, nil
}

func downloadSignedUrl(signed_url string, token string, timeout int) (*dicom.Dataset, error) {
	req, err := http.NewRequest("GET", signed_url, nil)
	if err != nil {
		return &dicom.Dataset{}, err
	}
	req.Header.Set("x-goog-meta-owner", token)
	req.Header.Set("Content-Type", "application/dicom")
	client := &http.Client{Timeout: time.Duration(timeout * 1e9)}
	resp, err := client.Do(req)
	if err != nil {
		return &dicom.Dataset{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &dicom.Dataset{}, &dicomweb.RequestError{StatusCode: resp.StatusCode, Err: errors.New(resp.Status)}
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &dicom.Dataset{}, err
	}
	return dicomutil.Bytes2Dicom(data)
}

func GetSignedUrl(api_url, study_instance_uid string, inference_command string, token string, timeout int, additional_params string) ([]*dicom.Dataset, error) {
	url := fmt.Sprintf(
		"%s/v3/studies/%s?inference_command=%s&signed_url=true%s",
		api_url,
		study_instance_uid,
		inference_command,
		additional_params,
	)
	headers := map[string]string{
		"x-goog-meta-owner": token,
		"Accept":            "application/json",
	}
	resp, err := dicomweb.Get(url, headers, timeout)
	if err != nil {
		if isFormattingError(err) {
			return []*dicom.Dataset{}, ErrFormattingError
		}
		return []*dicom.Dataset{}, err
	}
	defer resp.Body.Close()
	get_response := GetStudyResponseV3{}
	json.NewDecoder(resp.Body).Decode(&get_response)
	if get_response.SignedUrls == nil || len(*get_response.SignedUrls) == 0 {
		return []*dicom.Dataset{}, nil
	}
	dcm_slice := []*dicom.Dataset{}
	for _, signed_url := range *get_response.SignedUrls {
		dcm, err := downloadSignedUrl(signed_url, token, timeout)
		if err != nil {
			return []*dicom.Dataset{}, err
		}
		dcm_slice = append(dcm_slice, dcm)
	}
	get_response.SignedUrls = nil
	return dcm_slice, nil
}
func downloadSignedUrlToFile(signed_url string, token string, dcm_path string, timeout int) error {
	req, err := http.NewRequest("GET", signed_url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-goog-meta-owner", token)
	req.Header.Set("Content-Type", "application/dicom")
	client := &http.Client{Timeout: time.Duration(timeout * 1e9)}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &dicomweb.RequestError{StatusCode: resp.StatusCode, Err: errors.New(resp.Status)}
	}
	f, err := os.Create(dcm_path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func GetSignedUrlToFile(api_url, study_instance_uid string, inference_command string, token string, folder string, timeout int, additional_params string) ([]string, error) {
	res := []string{}
	url := fmt.Sprintf(
		"%s/v3/studies/%s?inference_command=%s&signed_url=true%s",
		api_url,
		study_instance_uid,
		inference_command,
		additional_params,
	)
	headers := map[string]string{
		"x-goog-meta-owner": token,
		"Accept":            "application/json",
	}
	resp, err := dicomweb.Get(url, headers, timeout)
	if err != nil {
		if isFormattingError(err) {
			return []string{}, ErrFormattingError
		}
		return []string{}, err
	}
	defer resp.Body.Close()
	get_response := GetStudyResponseV3{}
	json.NewDecoder(resp.Body).Decode(&get_response)
	if get_response.Status == "running" {
		return res, ErrPredictionRunning
	}
	if get_response.SignedUrls == nil || len(*get_response.SignedUrls) == 0 {
		return res, nil
	}
	for _, signed_url := range *get_response.SignedUrls {
		dcm_path := fmt.Sprintf("%s/%s", folder, dicomutil.RandomDicomName())
		err := downloadSignedUrlToFile(signed_url, token, dcm_path, timeout)
		if err != nil {
			return res, err
		}
		res = append(res, dcm_path)
	}
	return res, nil
}

func GetSmarturgences(api_url, study_instance_uid string, token string, timeout int) (GetSmarturgencesResponseV3, error) {
	url := fmt.Sprintf("%s/v3/smarturgences/%s", api_url, study_instance_uid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GetSmarturgencesResponseV3{}, err
	}
	req.Header.Set("x-goog-meta-owner", token)
	client := &http.Client{Timeout: time.Duration(timeout * 1e9)}
	resp, err := client.Do(req)
	if err != nil {
		return GetSmarturgencesResponseV3{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GetSmarturgencesResponseV3{}, &dicomweb.RequestError{StatusCode: resp.StatusCode, Err: errors.New(resp.Status)}
	}
	smarturgences_response := GetSmarturgencesResponseV3{}
	json.NewDecoder(resp.Body).Decode(&smarturgences_response)
	return smarturgences_response, nil
}

func GetSmartxpert(api_url, study_instance_uid string, token string, timeout int) (GetSmartxpertResponseV3, error) {
	url := fmt.Sprintf("%s/v3/smartxpert/%s", api_url, study_instance_uid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GetSmartxpertResponseV3{}, err
	}
	req.Header.Set("x-goog-meta-owner", token)
	client := &http.Client{Timeout: time.Duration(timeout * 1e9)}
	resp, err := client.Do(req)
	if err != nil {
		return GetSmartxpertResponseV3{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GetSmartxpertResponseV3{}, &dicomweb.RequestError{StatusCode: resp.StatusCode, Err: errors.New(resp.Status)}
	}
	smartxpert_response := GetSmartxpertResponseV3{}
	json.NewDecoder(resp.Body).Decode(&smartxpert_response)
	return smartxpert_response, nil
}
