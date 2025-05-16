package handlers

import (
	"LinkHUB/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// ImgurUploader Imgur图片上传实现
type ImgurUploader struct {
	ImgurApiUrl   string
	ImgurClientId string
}

// ImgurResponse Imgur API响应结构
type ImgurResponse struct {
	Data struct {
		Link       string `json:"link"`
		ID         string `json:"id"`
		DeleteHash string `json:"deletehash"`
	} `json:"data"`
	Success bool `json:"success"`
}

// NewImgurUploader 创建Imgur上传实例
func NewImgurUploader() *ImgurUploader {
	return &ImgurUploader{
		ImgurApiUrl:   config.GetConfig().Upload.ImgurApiUrl,
		ImgurClientId: config.GetConfig().Upload.ImgurClientId,
	}
}

// Upload 实现文件上传到Imgur
func (i *ImgurUploader) Upload(file *multipart.FileHeader) (string, string, error) {
	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", "", fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", i.ImgurApiUrl, bytes.NewReader(fileBytes))
	if err != nil {
		return "", "", fmt.Errorf("创建上传请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Client-ID "+i.ImgurClientId)
	req.Header.Set("Content-Type", "application/octet-stream")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("上传请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var imgurResp ImgurResponse
	if err := json.NewDecoder(resp.Body).Decode(&imgurResp); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}

	if !imgurResp.Success {
		return "", "", fmt.Errorf("上传到Imgur失败")
	}

	return imgurResp.Data.Link, imgurResp.Data.DeleteHash, nil
}

// Delete 实现文件删除
func (i *ImgurUploader) Delete(deleteHash string) error {
	// 创建DELETE请求
	deleteUrl := fmt.Sprintf("%s/%s", i.ImgurApiUrl, deleteHash)
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return fmt.Errorf("创建删除请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Client-ID "+i.ImgurClientId)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("删除请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("删除图片失败，状态码: %d", resp.StatusCode)
	}
	return nil
}
