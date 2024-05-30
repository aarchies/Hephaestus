package clickhousec

import (
	"context"
	"fmt"
)

// AsyncInsert 异步写入 本地表进行循环写入
func AsyncInsert(table string, data interface{}) error {

	cmd := fmt.Sprintf("INSERT INTO %s.%s SETTINGS async_insert=1, wait_for_async_insert=1", DataBase(), table)
	batch, err := DB().PrepareBatch(context.Background(), cmd)
	if err != nil {
		return err
	}

	if err := batch.AppendStruct(data); err != nil {
		return err
	}

	return batch.Send()
}
