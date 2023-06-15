# BlockChain_Go
A simple backend blockchain project in go programing language.
A part of Course Design of Software Engineering at Zhejiang University, 2022-2023, Group A4

### 技术栈

- Go
- Gin
- MongoDB

### feature

- 简单的区块链资产系统
- 路由操作 MongoDB 

### author

- https://github.com/HeartLinked

### Deploy to Linux

Enter the project directory and execute the following command to compile: 
`CGO_ ENABLED=0 GOOS=Linux GOARCH=amd64 go build main. go`

Upload the `main` file to any directory on Linux, execute `chmod 777 main` to increase permissions, and `nohup ./ Main &` Run the project.

The standard output and standard error output will be redirected to the `nohup.out` file in the current directory. If there is no `nohup.out` file in the current directory, this command will automatically create a new file and write the output to it. If the file already exists, this command will append the output to the end of the file. 
