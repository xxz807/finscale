package server

func NewServer() *gin.Engine {
	r := gin.New()

	// === 1. 基础保命设施 (必须有) ===
	r.Use(gin.Recovery()) // 防止崩溃
	r.Use(gin.Logger())   // 打印日志

	// === 2. 解决前端跨域 (必须有) ===
	// 简单粗暴允许所有跨域，开发阶段为了效率
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// === 3. 模拟鉴权 (为了让 Service 层有用户概念) ===
	// 以后这里换成真的 JWT 校验，现在先写死
	r.Use(func(c *gin.Context) {
		c.Set("userID", "1001") // 假装是管理员
		c.Next()
	})

	return r
}
