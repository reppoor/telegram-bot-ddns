package models

// Domain 用户数据模型
type Domain struct {
	ID               uint   `gorm:"primaryKey"`    // 主键
	Domain           string `gorm:"size:255"`      // 域名
	ForwardingDomain string `gorm:"size:255"`      // 转发域名
	IP               string `gorm:"size:255"`      // IP地址
	Port             int    `gorm:"size:255"`      // 端口
	ISP              string `gorm:"size:255"`      // 运营商
	Ban              bool   `gorm:"default:false"` // 是否启用
	Del              bool   `gorm:"default:false"` // 是否删除
	Weight           int    `gorm:"default:0"`     // 权重，默认值为0
	BanTime          int64  `gorm:"default:0"`     // 记录封禁时的 Unix 时间戳，0 表示未封禁或未记录
	SortOrder        int    `gorm:"default:0"`     // 排序权重
	RecordType       bool   `gorm:"default:true"`  // 是否为A记录
}

type TelegramPermission struct {
	ID         uint   `gorm:"primaryKey"`        // 主键
	TelegramID string `gorm:"size:255;not null"` // TelegramID
	IsAdmin    bool   `gorm:"default:false"`     //是否为管理员
	ban        bool   `gorm:"default:false"`     // 是否封禁

}
