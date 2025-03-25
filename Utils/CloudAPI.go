package Utils

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// CloudAPI Azure Blob Storage 操作客户端
type CloudAPI struct {
	connectionString string
	containerName    string
	BlobName         string
	containerClient  *container.Client
}

// NewCloudAPI 初始化 Azure Blob Storage 客户端
func NewCloudAPI() *CloudAPI {
	cfg := &CloudAPI{
		connectionString: "DefaultEndpointsProtocol=https;AccountName=httptestforlx;AccountKey=RWUgtSk0mz3tJhvqpHqk6pKvEQlwO7b/GphHXORn2HLzoL6VwXtX6yaqgqfIPt3b+RrLWVi/l/1e+AStSt1JQw==;EndpointSuffix=core.windows.net",
		containerName:    "java-functions-run-from-packages",
		BlobName:         "result.txt",
	}

	// 创建服务客户端
	serviceClient, err := azblob.NewClientFromConnectionString(cfg.connectionString, nil)

	// 获取容器客户端
	containerClient := serviceClient.ServiceClient().NewContainerClient(cfg.containerName)
	cfg.containerClient = containerClient

	// 检查并创建容器
	ctx := context.Background()
	_, err = containerClient.GetProperties(ctx, nil)
	if err != nil {
		containerClient.Create(ctx, nil)

	}
	return cfg
}

// UploadFile 上传文件到 Blob Storage
func (c *CloudAPI) UploadFile(localFilePath, blobName string) error {
	// 打开本地文件
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取 Blob 客户端
	if c.containerClient == nil {
		print("containerClient is nil\n")
	}
	blobClient := c.containerClient.NewBlockBlobClient(blobName)

	// 上传文件（强制覆盖）
	_, err = blobClient.UploadFile(context.Background(), file, &azblob.UploadFileOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr("application/octet-stream"),
		},
	})
	if err != nil {
		return fmt.Errorf("文件上传失败: %w", err)
	}

	return nil
}

// DownloadPartFile 下载指定范围的 Blob 内容
func (c *CloudAPI) DownloadPartFile(blobName string, startPos int, length int) ([]byte, error) {
	// 获取 Blob 客户端
	blobClient := c.containerClient.NewBlockBlobClient(blobName)
	
	// 设置下载范围（bytes=start-end）
	options := &azblob.DownloadStreamOptions{
		Range: azblob.HTTPRange{
			Offset: int64(startPos),
			Count:  int64(length),
		},
	}

	// 执行范围下载
	resp, err := blobClient.DownloadStream(context.Background(), options)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取内容到内存
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}

	return data, nil
}
