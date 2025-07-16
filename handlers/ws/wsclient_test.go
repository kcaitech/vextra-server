package ws

import (
	"testing"
)

// 模拟一个serve实现
type mockServe struct {
	closed bool
}

func (m *mockServe) handle(data *TransData, binaryData *([]byte)) {
	// 模拟处理
}

func (m *mockServe) close() {
	m.closed = true
}

// 测试isInterfaceNil函数
func TestIsInterfaceNil(t *testing.T) {
	// 测试真正的nil
	var nilServe ServeFace = nil
	if !isInterfaceNil(nilServe) {
		t.Error("isInterfaceNil应该返回true对于真正的nil")
	}

	// 测试非nil的serve
	mockServeInstance := &mockServe{}
	if isInterfaceNil(mockServeInstance) {
		t.Error("isInterfaceNil应该返回false对于非nil的serve")
	}

	// 测试接口类型有值但底层值为nil的情况
	var nilMockServe *mockServe = nil
	var interfaceServe ServeFace = nilMockServe

	// 直接比较nil会返回false，这是问题所在
	if interfaceServe == nil {
		t.Error("接口直接比较nil应该返回false（这是问题所在）")
	}

	// 但是isInterfaceNil应该能正确检测到
	if !isInterfaceNil(interfaceServe) {
		t.Error("isInterfaceNil应该能正确检测到底层值为nil的接口")
	}
}

// 测试bindServe方法对nil的处理
func TestBindServeNilHandling(t *testing.T) {
	// 创建一个模拟的WebSocket客户端
	client := &WSClient{
		serveMap: make(map[string]ServeFace),
	}

	// 测试绑定正常的serve
	mockServe1 := &mockServe{}
	client.bindServe("test1", mockServe1)

	if len(client.serveMap) != 1 {
		t.Errorf("绑定正常serve后，serveMap长度应该是1，实际是%d", len(client.serveMap))
	}

	// 测试绑定nil serve
	var nilMockServe *mockServe = nil
	var nilInterfaceServe ServeFace = nilMockServe

	client.bindServe("test2", nilInterfaceServe)

	if len(client.serveMap) != 1 {
		t.Errorf("绑定nil serve后，serveMap长度应该仍然是1，实际是%d", len(client.serveMap))
	}

	// 测试用nil serve替换已存在的serve
	client.bindServe("test1", nilInterfaceServe)

	if len(client.serveMap) != 0 {
		t.Errorf("用nil serve替换已存在的serve后，serveMap长度应该是0，实际是%d", len(client.serveMap))
	}

	// 验证原来的serve是否被正确关闭
	if !mockServe1.closed {
		t.Error("原来的serve应该被正确关闭")
	}
}

// 测试defer函数不会因为nil接口而panic
func TestDeferHandlerNoPanic(t *testing.T) {
	// 创建一个模拟的WebSocket客户端
	client := &WSClient{
		serveMap: make(map[string]ServeFace),
	}

	// 添加一个正常的serve
	mockServe1 := &mockServe{}
	client.serveMap["test1"] = mockServe1

	// 模拟原来的问题：直接存储nil接口
	var nilMockServe *mockServe = nil
	var nilInterfaceServe ServeFace = nilMockServe
	client.serveMap["test2"] = nilInterfaceServe

	// 测试defer函数是否会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("defer函数不应该panic，但是panic了：%v", r)
		}
	}()

	// 模拟defer函数中的逻辑
	for _, h := range client.serveMap {
		if h != nil {
			h.close() // 这里会panic，因为h是有类型但值为nil的接口
		}
	}
}

// 测试修复后的版本不会panic
func TestFixedVersionNoPanic(t *testing.T) {
	// 创建一个模拟的WebSocket客户端
	client := &WSClient{
		serveMap: make(map[string]ServeFace),
	}

	// 使用修复后的bindServe方法
	mockServe1 := &mockServe{}
	client.bindServe("test1", mockServe1)

	var nilMockServe *mockServe = nil
	var nilInterfaceServe ServeFace = nilMockServe
	client.bindServe("test2", nilInterfaceServe)

	// 测试defer函数是否会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("修复后的版本不应该panic，但是panic了：%v", r)
		}
	}()

	// 模拟defer函数中的逻辑
	for _, h := range client.serveMap {
		if h != nil {
			h.close()
		}
	}

	// 验证只有非nil的serve被关闭
	if !mockServe1.closed {
		t.Error("正常的serve应该被关闭")
	}
}

// 基准测试：比较直接nil检查和isInterfaceNil的性能
func BenchmarkDirectNilCheck(b *testing.B) {
	var nilMockServe *mockServe = nil
	var interfaceServe ServeFace = nilMockServe

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interfaceServe == nil
	}
}

func BenchmarkIsInterfaceNil(b *testing.B) {
	var nilMockServe *mockServe = nil
	var interfaceServe ServeFace = nilMockServe

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isInterfaceNil(interfaceServe)
	}
}
