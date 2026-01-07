package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderItem_TableName(t *testing.T) {
	tests := []struct {
		name     string
		orderItem OrderItem
		want     string
	}{
		{
			name:     "测试默认情况下返回正确的表名",
			orderItem: OrderItem{},
			want:     "order_items",
		},
		{
			name: "测试包含完整数据的OrderItem返回正确的表名",
			orderItem: OrderItem{
				ID:             1,
				OrderID:        1001,
				PlantID:        2001,
				SkuID:          3001,
				PlantName:      "测试植物",
				PlantLatinName: "Testus Plantus",
				SkuSize:        "Large",
				MainImgUrl:     "http://example.com/image.jpg",
				Price:          99.99,
				Quantity:       2,
			},
			want:     "order_items",
		},
		{
			name: "测试包含边界值的OrderItem返回正确的表名",
			orderItem: OrderItem{
				ID:             0,
				OrderID:        0,
				PlantID:        0,
				SkuID:          0,
				PlantName:      "",
				PlantLatinName: "",
				SkuSize:        "",
				MainImgUrl:     "",
				Price:          0.0,
				Quantity:       0,
			},
			want:     "order_items",
		},
		{
			name: "测试包含最大值的OrderItem返回正确的表名",
			orderItem: OrderItem{
				ID:             ^uint64(0), // 最大值
				OrderID:        ^uint64(0),
				PlantID:        ^uint64(0),
				SkuID:          ^uint64(0),
				PlantName:      "非常长的植物名称，用于测试边界情况，确保表名方法不会被影响",
				PlantLatinName: "Longissimus Plantarum Testiculorum Extremus",
				SkuSize:        "Extra Extra Large Plus Size",
				MainImgUrl:     "https://very-long-url.example.com/very/long/path/to/image.jpg?with=many&query=parameters&and=more",
				Price:          1.7976931348623157e+308, // float64 最大值
				Quantity:       ^uint(0),                 // uint 最大值
			},
			want:     "order_items",
		},
		{
			name: "测试包含特殊字符的OrderItem返回正确的表名",
			orderItem: OrderItem{
				PlantName:      "特殊名称!@#$%^&*()",
				PlantLatinName: "𝕋𝕖𝕤𝕥 𝕌𝕟𝕚𝕔𝕠𝕕𝕖 表情符号 🌟🎉",
				SkuSize:        "Size-特殊@字符",
				MainImgUrl:     "http://example.com/path?param=值&other=特殊",
			},
			want:     "order_items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.orderItem.TableName()
			require.Equal(t, tt.want, got, "TableName() = %v, want %v", got, tt.want)
		})
	}
}

// 基准测试，测试TableName方法的性能
func BenchmarkOrderItem_TableName(b *testing.B) {
	orderItem := OrderItem{
		ID:             1,
		OrderID:        1001,
		PlantID:        2001,
		SkuID:          3001,
		PlantName:      "基准测试植物",
		PlantLatinName: "Benchmarkus Plantus",
		SkuSize:        "Medium",
		MainImgUrl:     "http://example.com/benchmark.jpg",
		Price:          50.0,
		Quantity:       1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orderItem.TableName()
	}
}

// 示例测试，展示如何使用TableName方法
func ExampleOrderItem_TableName() {
	orderItem := OrderItem{
		PlantName: "示例植物",
		SkuSize:   "Small",
		Price:     29.99,
	}

	tableName := orderItem.TableName()
	println(tableName)
	// Output: order_items
}