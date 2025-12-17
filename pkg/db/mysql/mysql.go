package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql" // 引入MySQL驱动，实际为driver/mysql
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbInstances = make(map[string]*gorm.DB) // 存储GORM连接实例

func Init(mysqlConfigs map[string]MySQLConfig, initDb []string) error {
	for _, dbName := range initDb {
		cfg, ok := mysqlConfigs[dbName]
		if !ok {
			return fmt.Errorf("MySQL实例[%s]不存在于配置中", dbName)
		}
		// 构建DSN（Data Source Name）
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

		// 使用GORM打开连接
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("GORM连接失败[%s]: %v", dbName, err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("获取底层DB失败[%s]: %v", dbName, err)
		}

		// 设置连接池参数
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)                                  // 最大打开连接数
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)                                  // 最大空闲连接数
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.LifeTime) * time.Second) // 连接最大存活时间
		sqlDB.SetConnMaxIdleTime(300 * time.Second)                         // 连接最大空闲时间

		// 测试连接（真正建立连接）
		if err = sqlDB.Ping(); err != nil {
			return fmt.Errorf("连接测试失败[%s]: %v", dbName, err)
		}
		log.Printf("MySQL实例[%s]连接成功", dbName)
		dbInstances[dbName] = db
	}
	return nil
}

func GetDB(name string) (*gorm.DB, error) {
	db, exists := dbInstances[name]
	if !exists {
		return nil, fmt.Errorf("数据库实例[%s]不存在", name)
	}
	if db == nil {
		return nil, fmt.Errorf("数据库实例[%s]未初始化", name)
	}
	return db, nil
}

func Close() error {
	var errMsg string
	for name, db := range dbInstances {
		sqlDB, err := db.DB()
		if err != nil {
			errMsg += fmt.Sprintf("获取[%s]底层连接失败: %v; ", name, err)
			continue
		}
		if err := sqlDB.Close(); err != nil {
			errMsg += fmt.Sprintf("关闭[%s]失败: %v; ", name, err)
		}
	}
	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}
	return nil
}

// ExecuteSql 执行SQL并映射结果（利用GORM自动映射）
func ExecuteSql(dbName string, sqlStr string, args []interface{}, result interface{}) error {
	db, err := GetDB(dbName)
	if err != nil {
		return err
	}

	// 检查结果参数是否为有效指针
	resultVal := reflect.ValueOf(result)
	if resultVal.Kind() != reflect.Ptr || resultVal.IsNil() {
		return fmt.Errorf("result必须是非nil指针")
	}

	// 使用GORM执行原生SQL并自动映射结果
	if err := db.Raw(sqlStr, args...).Scan(result).Error; err != nil {
		// 保留原有的错误类型（如sql.ErrNoRows）
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("执行SQL失败: %v", err)
	}

	return nil
}

func scanStruct(rows *sql.Rows, structRes reflect.Value) error {
	fieldCount := structRes.NumField()           // 获取结构体字段数
	fieldPtrs := make([]interface{}, fieldCount) // 创建字段指针切片（用于Scan）
	for i := 0; i < fieldCount; i++ {
		fieldPtrs[i] = structRes.Field(i).Addr().Interface() // 获取字段指针
	}
	return rows.Scan(fieldPtrs...)
}

func scanSlice(rows *sql.Rows, sliceRes reflect.Value) error {
	elemType := sliceRes.Type().Elem() // 获取返回值的类型，如User{}
	for rows.Next() {
		elem := reflect.New(elemType).Elem()
		if err := scanStruct(rows, elem); err != nil {
			return err
		}
		sliceRes.Set(reflect.Append(sliceRes, elem))
	}
	return rows.Err()
}
