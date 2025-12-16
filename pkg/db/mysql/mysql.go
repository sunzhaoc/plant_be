package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql" // 引入MySQL驱动，实际为driver/mysql
)

var dbInstances = make(map[string]*sql.DB) // 存储已初始化的连接

func Init(mysqlConfigs map[string]MySQLConfig, initDb []string) error {
	for _, dbName := range initDb {
		cfg, ok := mysqlConfigs[dbName]
		if !ok {
			return fmt.Errorf("MySQL实例[%s]不存在于配置中", dbName)
		}
		// 构建DSN（Data Source Name）
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)
		// 打开数据库连接（不会立即建立连接，只是验证参数）
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("mysql connect failed: %v", err)
		}

		// 设置连接池参数
		db.SetMaxOpenConns(cfg.MaxOpen)                                  // 最大打开连接数
		db.SetMaxIdleConns(cfg.MaxIdle)                                  // 最大空闲连接数
		db.SetConnMaxLifetime(time.Duration(cfg.LifeTime) * time.Second) // 连接最大存活时间
		db.SetConnMaxIdleTime(300 * time.Second)                         // 连接最大空闲时间

		// 测试连接（真正建立连接）
		if err = db.Ping(); err != nil {
			log.Fatalf("mysql ping failed: %v", err)
		}
		log.Println("mysql connect success!")

		dbInstances[dbName] = db
	}
	return nil
}

func GetDB(name string) (*sql.DB, error) {
	db, exists := dbInstances[name]
	if !exists {
		return nil, fmt.Errorf("db %s not exists", name)
	}
	if db == nil {
		log.Panic("mysql connection pool is not initialized")
	}
	return db, nil
}

func Close() error {
	var errMsg string
	for name, db := range dbInstances {
		if err := db.Close(); err != nil {
			errMsg += fmt.Sprintf("关闭 %s 实例失败: %v; ", name, err)
		}
	}
	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}
	return nil
}

func ExecuteSql(dbName string, sqlStr string, args []interface{}, result interface{}) error {
	dbConn, err := GetDB(dbName)

	// 执行查询
	rows, err := dbConn.Query(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	// 检查结果参数是否为指针
	resultVal := reflect.ValueOf(result) // 返回 reflect.Value 对象
	if resultVal.Kind() != reflect.Ptr || resultVal.IsNil() {
		return fmt.Errorf("result 必须是非 nil 指针")
	}

	// 处理结果
	resultElem := resultVal.Elem() // 对指针类型的 reflect.Value 进行解引用，获取指针指向的实际值的 reflect.Value 对象
	switch resultElem.Kind() {
	case reflect.Slice: // 多条结果
		return scanSlice(rows, resultElem)
	case reflect.Struct: // 单条结果
		if rows.Next() {
			return scanStruct(rows, resultElem)
		}
		return sql.ErrNoRows
	default:
		return fmt.Errorf("result 仅支持结构指针或结构体切片指针")
	}
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
