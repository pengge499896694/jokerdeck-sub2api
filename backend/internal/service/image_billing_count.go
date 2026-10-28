package service

// resolveBillableImageCount 决定一次图片请求按几张计费。
//
// 只以响应体 data/output 中实际返回的图片为准。请求 n、HTTP 200 和分组
// 按张计费开关都不能证明出图成功；零产出不得兜底为一张或请求张数。
// 后两个参数保留调用兼容性，不再用作成功出图的证据。
func resolveBillableImageCount(reported, _ int, _ bool) int {
	if reported > 0 {
		return reported
	}
	return 0
}
