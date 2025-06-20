# File Management Service

Service used to centrally store the differnt type media files like images, video audio etc and can be accessible using the endpoints.
Initally it is going to support the Upload and Download functionality using `mongo DB grid-fs` and `go-kit` to access the resource.

## Service Flow

- Upload file workflow has the below steps:
    - Initialize the Upload process
        - Generate a session ID
        - Create a metadata in uploads collection
            - Chunk Size (defined by the client)
            - Total Chunks [file size/ divide by chunk size]
            - ID [Session ID unique]
            - Status : in-progress
            - Uploads Chunks: Used add the chunks ids which are processed
            - FileName
    - Upload Chunks
        - Chunks are store in the filesystem to under `tmp_uploads/sessionID` folders, just as a staging which can be used in case of partial completion.
        - Upload Chunks request consist of below files:
            - Chunk ID
            - Session ID
            - Data stream
        - Create the file with `chunkID.chunk` under the `tmp_uploads/sessionID`.
    - Complete Upload Status
        - Once the complete api called check metadata whether all the chunks are uploaded, if yes then upload the data into `grid-fs bucket`.
        - By default `grid-fs` stores file in to parts, `fs.chunks` and  `fs.files`.
        - `fs.chunks` stores main file contents below are the fields:
            - _id
            - file_id (unique id generated while upload the end data)
            - n (chunk id)
            - data (file data)
        - `fs.files` store the metadata about the file, which below fields:
            - _id
            - length [size of data uploaded in bytes]
            - chunk size (default is 261120 bytes or 255 KB)
            - UploadDate
            - fileName

- Download a file by name

## Service Usage

- Clone the repository
```
git clone https://github.com/ckshitij/file-mgmt-srv.git
```

- Intitialize the project
```
go mod tidy
```

- Change the config under `resoure/config.yml` like mongoDB URL.

- Run the service 
```
go run main.go
```

- Access the UI on ```http://localhost:8088/```

### Docker Compose 

- Directly run docker-compose it will boot up the below services:
    - mongo-db
    - mongo-express
        - ```http://localhost:8081/```
        - user: admin
        - password: pass
    - file-mgmt-srv

- Run docker-compose command
```
 docker-compose up --build -d
```

- Shutdown the docker-compose
```
docker-compose down
```