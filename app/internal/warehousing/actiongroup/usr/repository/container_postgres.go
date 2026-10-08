package repository

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/mrpostgres/builder/part"
	"github.com/mondegor/go-core/mrpostgres/db"
	"github.com/mondegor/go-core/mrstorage"

	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/enum/locationkind"
	"print-shop-back/internal/warehousing/module"
)

type (
	// ContainerPostgres - comment struct.
	ContainerPostgres struct {
		client          mrstorage.DBConnManager
		sqlBuilder      mrstorage.SQLConditionBuilder
		rowExistChecker db.RowExistsChecker[uint64]
	}
)

// NewContainerPostgres - создаёт объект ContainerPostgres.
func NewContainerPostgres(client mrstorage.DBConnManager) *ContainerPostgres {
	return &ContainerPostgres{
		client:     client,
		sqlBuilder: part.NewSQLConditionBuilder(),
		rowExistChecker: db.NewRowExistsChecker[uint64](
			client,
			module.DBTableNameContainers,
			"container_id",
			"",
		),
	}
}

// FetchByCondition - возвращает контейнеры аккаунта по фильтру, начиная за позицией params.Cursor
// (не более params.Cursor.Limit), и признак наличия записей за ней.
func (re *ContainerPostgres) FetchByCondition(ctx context.Context, params dto.ContainerParams) (rows []entity.Container, hasNext bool, err error) {
	condition := re.sqlBuilder.BuildFunc(
		func(c mrstorage.SQLConditionHelper) mrstorage.SQLPartFunc {
			conds := []mrstorage.SQLPartFunc{
				c.Equal("account_id", params.AccountID),
			}

			if params.Cursor.Code != "" {
				conds = append(
					conds,
					c.Expr("(container_code, container_marker) > (%s, %s)", params.Cursor.Code, params.Cursor.Marker),
				)
			}

			if params.Filter.SearchCode != "" {
				conds = append(conds, c.FilterEqual("container_code", params.Filter.SearchCode))
			}

			if len(params.Filter.SearchTags) > 0 {
				conds = append(conds, c.FilterInArray("container_tags", params.Filter.SearchTags))
			}

			return c.JoinAnd(conds...)
		},
	)

	whereStr, whereArgs := condition.ToSQL()

	sql := `
		SELECT
			container_id,
			tag_version,
			account_id,
			container_code,
			container_marker,
			container_volume,
			container_tags,
			container_images,
			created_at,
			updated_at
		FROM
			` + module.DBTableNameContainers + `
		WHERE
			` + whereStr + `
		ORDER BY
			account_id, container_code, container_marker
		FETCH FIRST ` + strconv.Itoa(params.Cursor.Limit+1) + ` ROWS ONLY;`

	cursor, err := re.client.Conn(ctx).Query(
		ctx,
		sql,
		whereArgs...,
	)
	if err != nil {
		return nil, false, err
	}

	defer cursor.Close()

	for cursor.Next() {
		// лишняя строка сверх limit лишь сообщает, что за текущей страницей есть записи
		if len(rows) == params.Cursor.Limit {
			hasNext = true

			break
		}

		var row entity.Container

		err = cursor.Scan(
			&row.ID,
			&row.TagVersion,
			&row.AccountID,
			&row.Code,
			&row.Marker,
			&row.Volume,
			&row.Tags,
			&row.Images,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, false, err
		}

		row.Kind = locationkind.ByID(row.ID)

		if rows == nil {
			rows = make([]entity.Container, 0, params.Cursor.Limit)
		}

		rows = append(rows, row)
	}

	return rows, hasNext, cursor.Err()
}

// FetchOne - comment method.
func (re *ContainerPostgres) FetchOne(ctx context.Context, accountID uuid.UUID, rowID uint64) (row entity.Container, err error) {
	sql := `
		SELECT
			container_id,
			tag_version,
			account_id,
			container_code,
			container_marker,
			container_volume,
			container_tags,
			container_images,
			created_at,
			updated_at
		FROM
			` + module.DBTableNameContainers + `
		WHERE
			container_id = $1 AND account_id = $2
		FETCH FIRST 1 ROW ONLY;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
	).Scan(
		&row.ID,
		&row.TagVersion,
		&row.AccountID,
		&row.Code,
		&row.Marker,
		&row.Volume,
		&row.Tags,
		&row.Images,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	row.Kind = locationkind.ByID(row.ID)

	return row, err
}

// IsExist - comment method.
// result: nil - exists, errors.ErrEventStorageNoRecordFound - not exists, error - query error
func (re *ContainerPostgres) IsExist(ctx context.Context, accountID uuid.UUID, rowID uint64) error {
	sql := `
		SELECT
            1
        FROM
            ` + module.DBTableNameContainers + `
        WHERE
            container_id = $1 AND account_id = $2
        FETCH FIRST 1 ROW ONLY;`

	var value uint64

	err := re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
	).Scan(
		&value,
	)

	return err
}

// FetchMaxMarker - comment method.
func (re *ContainerPostgres) FetchMaxMarker(ctx context.Context, accountID uuid.UUID, code string) (marker uint16, err error) {
	sql := `
		SELECT
			MAX(container_marker)
		FROM
			` + module.DBTableNameContainers + `
		WHERE
			account_id = $1 AND container_code = $2
		GROUP BY
			account_id, container_code
		FETCH FIRST 1 ROW ONLY;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		accountID,
		code,
	).Scan(
		&marker,
	)

	return marker, err
}

// Insert - comment method.
func (re *ContainerPostgres) Insert(ctx context.Context, row entity.Container) (rowID uint64, err error) {
	sql := `
		INSERT INTO ` + module.DBTableNameContainers + `
			(
				container_id,
				account_id,
				container_code,
				container_marker,
				container_volume,
				container_tags,
				container_images
			)
		VALUES
			(nextval('` + row.SequenceName() + `'), $1, $2, $3, $4, $5, $6)
		RETURNING
            container_id;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		row.AccountID,
		row.Code,
		row.Marker,
		row.Volume,
		row.Tags,
		row.Images,
	).Scan(
		&rowID,
	)

	return rowID, err
}

// UpdateTags - comment method.
func (re *ContainerPostgres) UpdateTags(ctx context.Context, row entity.UpdateContainerTags) (tagVersion uint32, err error) {
	sql := `
		UPDATE
			` + module.DBTableNameContainers + `
		SET
			tag_version = tag_version + 1,
			container_tags = $4,
			updated_at = NOW()
		WHERE
			container_id = $1 AND tag_version = $2 AND account_id = $3
		RETURNING
            tag_version;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		row.ID,
		row.TagVersion,
		row.AccountID,
		row.Tags,
	).Scan(
		&tagVersion,
	)

	return tagVersion, err
}
