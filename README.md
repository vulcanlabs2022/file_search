# File Search
## Install
Build Docker Image:

```sh
docker build -t calehh/file_search:dev .
```

拉取镜像
```sh
docker pull calehh/file_search:dev
```
https://hub.docker.com/repository/docker/calehh/file_search

## Startup

编排参考docker-compose-example.yml

```sh
docker compose up
```

## API
### Host
http://localhost:6317

### 返回结构

返回示例：{"code":0, "data":"XXX"}

code为0则成功，data为返回信息。小于0为错误，data为错误信息。

### 添加文件 http://127.0.0.1:6317/api/file/input

#### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

#### 请求字段：

| 请求字段 | 类型     | 备注                             |
| -------- | -------- | -------------------------------- |
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

### 删除文件 http://127.0.0.1:6317/api/file/delete

#### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

#### 请求字段：

| 请求字段 | 类型   | 备注           |
| -------- | ------ | -------------- |
| docId    | string | 文件编号 DocID |

#### 返回：

```
{
   code: 0,
   data : "5c6390bb-abc4-41c1-8e97-8215fe74a066" //删除文件的编号DocID
}
```

### 查找文件 http://127.0.0.1:6317/api/file/query

#### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

#### 请求字段：

| 请求字段 | 类型   | 备注                          |
| -------- | ------ | ----------------------------- |
| query    | string | 查询文本                      |
| limit    | int    | 最大回复数 （暂时不支持分页） |

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
              index: 'Files', //索引名，为Files或Rss
              name: 'aaa.js', //文件名
              docId: "5c6390bb-abc4-41c1-8e97-8215fe74a066", //文件编号
              where: "/131313/bbb", //路径
              type: ".js", //扩展名
              size: number, //字节数
              created : number, //创建时间戳
              snippet: string, //高亮摘要，用<mark>标签标注 例如：…and the second-smallest planet in the <mark>Solar</mark> <mark>System</mark>, larger only than Mercury. In the English language, Mars is named for the Roman god of war. Mars is a terrestrial planet with a thin atmosphere and h…
         }
    ]
   }
}
```

### 添加RSS http://127.0.0.1:6317/api/rss/input

#### 请求格式
Post Json

```
{
   name: 'aaa',
   entry_id : number,
   created : number,
   feed_infos : [{        // 属于哪个feed列表
      feed_id : number,
      feed_name : number;
      feed_icon : string;
   }],
   borders: [          // 属于哪个文件夹
      { 
                     name:string;
                     id: number;
      }
   ],
    "content": string, //rss内容
}
```

#### 返回：

```
{
   code: 0,
   data : "5c6390bb-abc4-41c1-8e97-8215fe74a066" //添加文件的编号DocID，DocID对应唯一文件
}
```


### 删除Rss http://127.0.0.1:6317/api/rss/delete

#### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

#### 请求字段：

| 请求字段 | 类型   | 备注           |
| -------- | ------ | -------------- |
| docId    | string | 文件编号 DocID |

#### 返回：

```
{
   code: 0,
   data : "5c6390bb-abc4-41c1-8e97-8215fe74a066" //删除文件的编号DocID
}
```


### 查找Rss http://127.0.0.1:6317/api/rss/query

#### 请求格式
post请求使用表单格式

Content-Type:multipart/form-data

#### 请求字段：

| 请求字段 | 类型   | 备注                          |
| -------- | ------ | ----------------------------- |
| query    | string | 查询文本                      |
| limit    | int    | 最大回复数 （暂时不支持分页） |

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
              name: 'aaa',
              entry_id : number,
              created : number,
              feed_infos : [{        // 属于哪个feed列表
                  feed_id : number,
                  feed_name : number;
                  feed_icon : string;
              }],
              borders: [          // 属于哪个文件夹
                  { 
                     name:string;
                     id: number;
                  }
              ],
              docId: string,
              snippet：'' ， // 如果是文件是文字的，显示命中了哪段话
         }
    ]
   }
}
```