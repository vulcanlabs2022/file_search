# File Search
## Install
Build Docker Image:

```sh
docker build -t calehh/file_search:latest .
```

拉取镜像
```sh
docker pull calehh/file_search:latest
```
https://hub.docker.com/repository/docker/calehh/file_search

## Startup
```sh
docker run -p 6317:6317 -e ZINC_FIRST_ADMIN_USER=admin -e ZINC_FIRST_ADMIN_PASSWORD=User#123 -e ZINC_URI="http://host.docker.internal:4080" --name searcher calehh/file_search:latest
```
ZINC_FIRST_ADMIN_USER，ZINC_FIRST_ADMIN_PASSWORD分别为zincsearch的用户名和密码。ZINC_URI为zincsearch服务URI。

## API
### Host
http://localhost:6317

### 索引和文件
filesearch可以建立多个索引，通过索引名index标识。添加、删除和查询文件将在指定的某个索引内进行，如果没有指定则使用默认名为terminus的索引。

### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

### 返回结构

返回示例：{"code":0, "data":"XXX"}

code为0则成功，data为返回信息。小于0为错误，data为错误信息。

### 添加文件 http://127.0.0.1:6317/api/input

#### 请求字段：

| 请求字段 | 类型     | 备注                             |
| -------- | -------- | -------------------------------- |
| index    | string   | 索引名（可选，默认为Files）  |
| doc      | file文件 | 文件上传                         |
| path     | string   | 文件路径                         |
| filename | string   | 文件名（可选，默认为上传文件名） |
| content  | string   | 文本内容（可选，暂时备用）       |

文件内容来自doc或者content二选一，doc优先级更高。

#### 返回：

```
{
   code: 0,
   data : "5c6390bb-abc4-41c1-8e97-8215fe74a066" //添加文件的编号DocID，DocID对应唯一文件
}
```

### 删除文件 http://127.0.0.1:6317/api/delete

#### 请求字段：

| 请求字段 | 类型   | 备注                            |
| -------- | ------ | ------------------------------- |
| index    | string | 索引名（可选，默认为Files） |
| docId    | string | 文件编号 DocID                  |

#### 返回：

```
{
   code: 0,
   data : "5c6390bb-abc4-41c1-8e97-8215fe74a066" //删除文件的编号DocID
}
```

### 查找文件 http://127.0.0.1:6317/api/query

#### 请求字段：

| 请求字段 | 类型   | 备注                            |
| -------- | ------ | ------------------------------- |
| index    | string | 索引名（可选，默认为Files） |
| query    | string | 查询文本                        |
| limit    | int    | 最大回复数 （暂时不支持分页）   |

#### 返回：

```
{
   code: 0
   data : {
     count: 10,
     offset : 0,
     limit : 10,
     items: [
        {	
              name: 'aaa.js', //文件名
              docId: "5c6390bb-abc4-41c1-8e97-8215fe74a066", //文件编号
              where: "/131313/bbb", //路径
              type: "js", //扩展名
              size: number, //字节数
              created : number, //创建时间戳
              content：'' ， //文件内容
         }
    ]
   }
}
```