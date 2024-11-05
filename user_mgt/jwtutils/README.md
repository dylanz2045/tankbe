## 函数的解释、作用

### 对于内部函数
func VerifyToken(token *jwt.Token, auth string) (bool, error)
返回值有三个
- bool 就是超时
- err 是否有错误