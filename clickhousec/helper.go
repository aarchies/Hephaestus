package clickhousec

import (
	"context"
	"fmt"
)

// AsyncInsert 异步写入
func AsyncInsert(table string, data interface{}) error {

	batch, err := DB().PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %s.%s SETTINGS async_insert=1, wait_for_async_insert=1", DataBase(), table))
	if err != nil {
		return err
	}

	if err := batch.AppendStruct(data); err != nil {
		return err
	}

	return batch.Send()
}
