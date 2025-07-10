package u

import (
	"encoding/json"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Error(code int, msg string) map[string]interface{} {
	return map[string]interface{}{"code": code, "msg": msg, "success": false}
}
func OK(code int, msg string) map[string]interface{} {
	return map[string]interface{}{"code": code, "msg": msg}
}

func OKK(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	if json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "msg": "成功"}) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
func Sucess(code int, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"code": code, "data": data}
}
func SucessWithData(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"code": 0, "data": data}
}
func SucessWithObject(data interface{}) map[string]interface{} {
	return map[string]interface{}{"code": 0, "data": data}
}
func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if json.NewEncoder(w).Encode(data) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	//w.WriteHeader(400)
}

func RespondObject(w http.ResponseWriter, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if json.NewEncoder(w).Encode(data) != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetDataByJson[T any](r *http.Request) (*T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ClearTemp() error {
	tempDir := glog.TempDir()
	glog.Debug(tempDir)
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}
	for _, entry := range entries {
		fullPath := filepath.Join(tempDir, entry.Name())
		err = os.RemoveAll(fullPath)
		if err != nil {
			glog.Errorf("删除失败 %s  %v", fullPath, err)
		} else {
			glog.Debugf("删除成功 %s", fullPath)
		}
	}
	return err
}

func IsMacOs() bool {
	if strings.Compare(runtime.GOOS, "darwin") == 0 {
		return true
	}
	return false
}
