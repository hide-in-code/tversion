# tversion
一个纯go编写的mini型版本控制工具

# 编译

    git clone https://github.com/hide-in-code/tversion.git

    cd tversion

    go build -o tver main.go

    cp tver /usr/local/bin/ 或者使用alias

# 使用

    cd /path/to/your/verdata

### 提交版本
    tver commit

### 查看所有版本
    tver versions

### 回退版本
    tver checkout <commitId>

# 协议
MIT License