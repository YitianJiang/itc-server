# itc-server

## 运行程序

```bash
cd ~/go/src/code.byted.org/clientQA/itc-server
bash run.sh
```

## 加入我们

### 拉取代码

```Bash
mkdir -p ~/go/src/code.byted.org/clientQA
cd ~/go/src/code.byted.org/clientQA
git clone git@code.byted.org:clientQA/itc-server.git
```

## 一些建议

使用数据库时，请采用`database.DB()`的方式，该方法会返回一个全局的数据库句柄，全局数据库句柄会在`main`函数中初始化。采用这种方式有两个好处：

- 减少代码冗余，数据库句柄仅初始化一次，全局可用。

- 数据库句柄限制为包级可见，可避免不知情的修改，提高代码的健壮性。