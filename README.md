一、项目导入与依赖管理
1、本代码是基于Azure平台的Go代码实现，将该代码导入支持Go的IDE中。
2、根据go.sum自动导入一些需要的库和包（GO版本）。

二、云服务准备
1、准备一个Microsoft Azure账户（腾讯云账户），确保有可用的订阅余额
2、创建一个云函数服务（Azure Function）和存储服务（Azure Blob Storage），具体创建可以根据官方文档。其中云函数创建时需要选择类型为Http触发的云函数。

三、云部署和配置
1、以IDEA中部署云函数为例，下载Azure Tool官方插件，具体流程为在 IDEA 项目资源管理器中选择 Azure 图标，然后选择“部署到 Azure -> 部署到 Azure Functions” ，接下来的选项都为默认设 置即可，最后点击“运行”就会自动部署函数。
2、替换 Client.go中的云函数URL为上面创建的云函数的URL以触发云函数。替换CloudAPI.go中的connectionString为BLOB存储的密钥、containerName为BLOB容器的名字、BlobName为具体文件的名字。

四、系统运行
1、替换必要参数：filePath为需要存储到云端的数据的地址。
2、启动Clien“t.go代码，按照需要输入BLOCK_SHARDS和DATA_SHARDS，系统开始运行。

