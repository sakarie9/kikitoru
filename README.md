# Kikitoru

一个同人音声专用的音乐流媒体服务器

**！本项目仍在早期开发阶段，可能会出现预期之外的问题或 API 变更！**

### 特点
- 重写的 [kikoeru](https://github.com/kikoeru-project/kikoeru-express)，尽量保持 API 不变，可以兼容原版前端及第三方 APP 的大部分功能
- 使用 Golang 编写，性能和稳定性更高
- 数据库使用 Postgresql
- 扫描速度快，900 部作品建立数据库并下载封面仅需 1min15s 
- 内存占用低 ~22M

### 功能介绍
- 从 DLSite 爬取音声元数据
- 对音声标记进度、打星、写评语
- 通过标签或关键字快速检索想要找到的音声
- 根据音声元数据对检索结果进行排序
- 支持在 Web 端修改配置文件和扫描音声库
- 支持为音声库添加多个根文件夹

### 部署

<details>
<summary>
<b>Docker Compose</b>
</summary>

1. 下载 [docker-compose.yaml](docker-compose.yaml) 到任意目录下
2. 新建 `data` 文件夹
3. 下载 [kikitoru-quasar/spa.zip](https://github.com/sakarie9/kikitoru-quasar/releases/latest/spa.zip)，解压到 `/data/dist` 下，保证存在 `/data/dist/index.html`

    正确配置后目录结构应如下
    
    ```shell
    kikitoru
    ├── data
    │   ├── config.json # 配置文件，运行后生成
    │   ├── dist
    │   │   ├── css
    │   │   ├── fonts
    │   │   ├── index.html
    │   │   ├── js
    │   │   └── statics
    │   └── postgresql # 数据库文件夹
    └── docker-compose.yaml
    ```
   
4. 在 `docker-compose.yaml` 的同级目录下运行 `docker-compose up -d` 启动

</details>
