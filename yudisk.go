package YuDisk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ydError struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

const YandexDiskApiEndpoint = "https://cloud-api.yandex.net"
const CurrentUserAgent = "Mozilla/5.0 (Linux; Android 13; SM-A037U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36"

type YDApi struct {
	*http.Request
	*http.Transport
}

func buildError(funcName string, message string, wrap error) error {
	return fmt.Errorf("[%s] %s: %w", funcName, message, wrap)
}

func (yde *ydError) getYDError(resp *http.Response) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return buildError("getYDError", "fail read body", err)
	}
	err = json.Unmarshal(raw, yde)
	if err != nil {
		return buildError("getYDError", "fail Unmarshal", err)
	}
	return nil
}

func parseStatus(resp *http.Response) error {
	ydErrorObj := ydError{}
	var rError error = nil
	// todo: add full code parse
	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 && resp.StatusCode != 204 {
		err := ydErrorObj.getYDError(resp)
		if err != nil {
			rError = buildError("parseStatus", fmt.Sprintf("fail get YD error struct, response status code - %d", resp.StatusCode), err)
		} else {
			rError = errors.New(fmt.Sprintf("%+v", ydErrorObj))
		}
	}
	return rError
}

func (yd YDApi) SetUserAgent(agent string) {
	yd.Header.Set("User-Agent", agent)
}

func (yd YDApi) getResponseBodyBytes() ([]byte, error) {
	clt := http.Client{Transport: yd.Transport}
	resp, err := clt.Do(yd.Request)
	if err != nil {
		return nil, buildError("getResponseBodyBytes", "fail do request", err)
	}
	defer resp.Body.Close()
	errStatus := parseStatus(resp)
	if errStatus != nil {
		return nil, buildError("getResponseBodyBytes", "fail get response status", errStatus)
	}
	raw, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return nil, buildError("getResponseBodyBytes", "fail read response body", errRead)

	}
	return raw, nil
}

func (yd YDApi) SetProxy(proxyUrl string) error {
	Url, err := url.Parse(proxyUrl)
	if err != nil {
		return buildError("SetProxy", "fail parse url string", err)
	} else {
		yd.Proxy = http.ProxyURL(Url)
	}
	return nil
}

func (yd YDApi) setUrl(newUrl string) error {
	Url, err := url.Parse(newUrl)
	if err != nil {
		return buildError("setUrl", "fail parse url string", err)
	} else {
		yd.URL = Url
		yd.Host = Url.Host
	}
	return nil
}

func (yd YDApi) GetResourceMeta(path string) (Resource, error) {
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources?path=" + path)
	yd.Method = http.MethodGet
	if errUrl != nil {
		return Resource{}, buildError("GetResourceMeta", "fail set url", errUrl)
	}
	raw, err := yd.getResponseBodyBytes()
	if err != nil {
		return Resource{}, buildError("GetResourceMeta", "fail request", err)
	}
	metaResource := Resource{}
	metaResource.Embedded = ResourceList{}
	metaResource.Embedded.Items = make([]ResourceItem, 0)
	err = json.Unmarshal(raw, &metaResource)
	if err != nil {
		return Resource{}, fmt.Errorf("fail unmarshal Resource stcruct: %w", err)
	}
	return metaResource, nil
}

func (yd YDApi) OperationStatus(operationID string) (string, error) {
	OperationStatus := struct {
		Status string `json:"status"`
	}{}
	yd.Method = http.MethodGet
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/operations/" + operationID)
	if errUrl != nil {
		return "", buildError("OperationStatus", "fail set url", errUrl)
	}
	rawResp, errResp := yd.getResponseBodyBytes()
	if errResp != nil {
		return "", buildError("OperationStatus", "fail request", errResp)
	}
	err := json.Unmarshal(rawResp, &OperationStatus)
	if err != nil {
		return "", buildError("OperationStatus", "fail Unmarshal", err)
	}
	return OperationStatus.Status, nil
}

func (yd YDApi) uploadFileOperation(path string, operation Operation) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return buildError("uploadFileOperation", "fail read file", err)
	}
	uploadRequest, errRequest := http.NewRequest(operation.Method, operation.Href, bytes.NewBuffer(raw))
	if errRequest != nil {
		return buildError("uploadFileOperation", "fail build new request", errRequest)
	}
	uploadRequest.Header.Set("User-Agent", yd.UserAgent())
	yd.Request = uploadRequest
	_, errResp := yd.getResponseBodyBytes()
	if errResp != nil {
		return buildError("uploadFileOperation", "fail request", err)
	}
	return nil
}

func (yd YDApi) Upload(pathLocal string, pathCloud string, overwrite bool) (string, error) {
	yd.Method = http.MethodGet
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources/upload?path=" + pathCloud + "&overwrite=" + fmt.Sprintf("%t", overwrite))
	if errUrl != nil {
		return "", buildError("Upload", "fail set url", errUrl)
	}
	raw, err := yd.getResponseBodyBytes()
	if err != nil {
		return "", buildError("Upload", "fail request", err)
	}
	operationUpload := Operation{}
	err = json.Unmarshal(raw, &operationUpload)
	if err != nil {
		return "", buildError("Upload", "fail Unmarshal", err)
	}
	if !operationUpload.Templated {
		err := yd.uploadFileOperation(pathLocal, operationUpload)
		if err != nil {
			return "", buildError("Upload", "fail do upload operation", err)
		}
	} else {
		// todo: add check template
		return "", buildError("Download", "operation.Template = true", nil)
	}
	return operationUpload.OperationId, nil
}

func (yd YDApi) downloadFileOperation(operation Operation) ([]byte, error) {
	yd.Method = operation.Method
	errUrl := yd.setUrl(operation.Href)
	if errUrl != nil {
		return nil, buildError("downloadFileOperation", "fail set url", errUrl)
	}
	// todo: dont use getResponseBodyBytes
	raw, err := yd.getResponseBodyBytes()
	if err != nil {
		return nil, buildError("downloadFileOperation", "fail request", err)
	}
	return raw, nil
}

func (yd YDApi) Download(path string) ([]byte, error) {
	var fileContent []byte
	yd.Method = http.MethodGet
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources/download?path=" + path)
	if errUrl != nil {
		return nil, buildError("Download", "fail set url", errUrl)
	}
	raw, err := yd.getResponseBodyBytes()
	if err != nil {
		return nil, buildError("Download", "fail request", err)
	}
	operationDownload := Operation{}
	err = json.Unmarshal(raw, &operationDownload)
	if err != nil {
		return nil, buildError("Download", "fail Unmarshal", err)
	}
	if !operationDownload.Templated {
		fileContent, err = yd.downloadFileOperation(operationDownload)
		if err != nil {
			return nil, buildError("Download", "fail do download operation", err)
		}
	} else {
		// todo: add check template
		return nil, buildError("Download", "operation.Template = true", nil)
	}
	return fileContent, nil
}

func (yd YDApi) MkDir(path string) error {
	yd.Method = http.MethodPut
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources?path=" + path)
	if errUrl != nil {
		return buildError("MkDir", "fail set url", errUrl)
	}
	_, err := yd.getResponseBodyBytes()
	if err != nil {
		return buildError("MkDir", "fail request", err)
	}
	// todo: parse response json structure
	return nil
}

func (yd YDApi) Copy(from string, to string, overwrite bool) error {
	yd.Method = http.MethodPost
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources/copy?from=" + from + "&path=" + to + "&overwrite=" + fmt.Sprintf("%t", overwrite))
	if errUrl != nil {
		return buildError("Copy", "fail set url", errUrl)
	}
	_, err := yd.getResponseBodyBytes()
	if err != nil {
		return buildError("Copy", "fail request", err)
	}
	return nil
}

func (yd YDApi) Move(from string, to string, overwrite bool) error {
	yd.Method = http.MethodPost
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources/move?from=" + from + "&path=" + to + "&overwrite=" + fmt.Sprintf("%t", overwrite))
	if errUrl != nil {
		return buildError("Move", "fail set url", errUrl)
	}
	_, err := yd.getResponseBodyBytes()
	if err != nil {
		return buildError("Move", "fail request", err)
	}
	return nil
}

func (yd YDApi) Delete(pathToCloudFile string, permanently bool) error {
	errUrl := yd.setUrl(YandexDiskApiEndpoint + "/v1/disk/resources?path=" + pathToCloudFile + "&permanently=" + fmt.Sprintf("%t", permanently))
	if errUrl != nil {
		return buildError("Delete", "fail set url", errUrl)
	}
	yd.Method = http.MethodDelete
	_, err := yd.getResponseBodyBytes()
	if err != nil {
		return buildError("Delete", "fail request", err)
	}
	return nil
}

func NewYuDisk(token string) (YDApi, error) {
	request, err := http.NewRequest(http.MethodGet, YandexDiskApiEndpoint, nil)
	request.Header.Set("User-Agent", CurrentUserAgent)
	request.Header.Set("Authorization", "OAuth "+token)
	if err != nil {
		return YDApi{}, buildError("NewYuDisk", "fail create new request", err)
	}
	return YDApi{request, &http.Transport{}}, nil
}
