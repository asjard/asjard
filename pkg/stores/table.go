package stores

type Table interface {
	// 表名，或者collection名称
	TableName() string
	// 数据库名称
	Database() string
}
