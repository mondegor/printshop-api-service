package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/mrpostgres/builder/helper"
	"github.com/mondegor/go-core/mrpostgres/db"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"

	"print-shop-back/internal/controls/submitform/module"
	"print-shop-back/internal/controls/submitform/section/adm/entity"
)

type (
	// FormElementPostgres - comment struct.
	FormElementPostgres struct {
		client             mrstorage.DBConnManager
		sqlBuilder         mrstorage.SQLBuilder
		sqlConditionHelper mrstorage.SQLConditionHelper
		rowExistChecker    db.RowExistsChecker[uint64]
		repoSoftDeleter    db.RowSoftDeleter[uint64]
	}
)

// NewFormElementPostgres - создаёт объект FormElementPostgres.
func NewFormElementPostgres(client mrstorage.DBConnManager, sqlBuilder mrstorage.SQLBuilder) *FormElementPostgres {
	return &FormElementPostgres{
		client:             client,
		sqlBuilder:         sqlBuilder,
		sqlConditionHelper: helper.NewSQLCondition(),
		rowExistChecker: db.NewRowExistsChecker[uint64](
			client,
			module.DBTableNameSubmitFormElements,
			"element_id",
			module.DBFieldDeletedAt,
		),
		repoSoftDeleter: db.NewRowSoftDeleter[uint64](
			client,
			module.DBTableNameSubmitFormElements,
			"element_id",
			module.DBFieldTagVersion,
			module.DBFieldDeletedAt,
		),
	}
}

// NewCondition - возвращает условие списка сортировки: неудалённые элементы указанной формы.
func (re *FormElementPostgres) NewCondition(formID uuid.UUID) mrstorage.SQLPartFunc {
	return re.sqlConditionHelper.JoinAnd(
		re.sqlConditionHelper.Equal("form_id", formID),
		re.sqlConditionHelper.Expr("deleted_at IS NULL"),
	)
}

// Fetch - comment method.
func (re *FormElementPostgres) Fetch(ctx context.Context, formID uuid.UUID) ([]entity.FormElement, error) {
	sql := `
        SELECT
            fe.element_id,
            fe.tag_version,
			fe.form_id,
            fe.param_name,
            fe.element_caption,
			fe.template_id,
			fe.template_version,
			fe.element_required,
            et.element_type,
            et.element_detailing,
			et.element_body,
            fe.created_at,
			fe.updated_at
        FROM
            ` + module.DBTableNameSubmitFormElements + ` fe
        JOIN
            ` + module.DBTableNameElementTemplates + ` et
        ON
            fe.template_id = et.template_id
        WHERE
            fe.form_id = $1 AND fe.deleted_at IS NULL
        ORDER BY
            fe.order_index ASC, fe.element_id ASC;`

	cursor, err := re.client.Conn(ctx).Query(
		ctx,
		sql,
		formID,
	)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	rows := make([]entity.FormElement, 0)

	for cursor.Next() {
		var row entity.FormElement

		err = cursor.Scan(
			&row.ID,
			&row.TagVersion,
			&row.FormID,
			&row.ParamName,
			&row.Caption,
			&row.TemplateID,
			&row.TemplateVersion,
			&row.Required,
			&row.Type,
			&row.Detailing,
			&row.Body,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, cursor.Err()
}

// FetchOne - comment method.
func (re *FormElementPostgres) FetchOne(ctx context.Context, rowID uint64) (entity.FormElement, error) {
	sql := `
        SELECT
            fe.tag_version,
			fe.form_id,
            fe.param_name,
            fe.element_caption,
			fe.template_id,
			fe.template_version,
			fe.element_required,
            et.element_type,
            et.element_detailing,
			et.element_body,
            fe.created_at,
			fe.updated_at
        FROM
            ` + module.DBTableNameSubmitFormElements + ` fe
        JOIN
            ` + module.DBTableNameElementTemplates + ` et
        ON
            fe.template_id = et.template_id
        WHERE
            fe.element_id = $1 AND fe.deleted_at IS NULL
        LIMIT 1;`

	row := entity.FormElement{ID: rowID}

	err := re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
	).Scan(
		&row.TagVersion,
		&row.FormID,
		&row.ParamName,
		&row.Caption,
		&row.TemplateID,
		&row.TemplateVersion,
		&row.Required,
		&row.Type,
		&row.Detailing,
		&row.Body,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	return row, err
}

// FetchIDByParamName - comment method.
func (re *FormElementPostgres) FetchIDByParamName(ctx context.Context, formID uuid.UUID, paramName string) (rowID uint64, err error) {
	sql := `
        SELECT
            element_id
        FROM
            ` + module.DBTableNameSubmitFormElements + `
        WHERE
            form_id = $1 AND param_name = $2 AND deleted_at IS NULL
        LIMIT 1;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		formID,
		paramName,
	).Scan(
		&rowID,
	)

	return rowID, err
}

// IsExist - comment method.
// result: nil - exists, errors.ErrEventStorageNoRecordFound - not exists, error - query error
func (re *FormElementPostgres) IsExist(ctx context.Context, rowID uint64) error {
	return re.rowExistChecker.IsExist(ctx, rowID)
}

// Insert - comment method.
func (re *FormElementPostgres) Insert(ctx context.Context, row entity.FormElement) (rowID uint64, err error) {
	sql := `
        INSERT INTO ` + module.DBTableNameSubmitFormElements + `
            (
                form_id,
                param_name,
                element_caption,
				template_id,
				template_version,
                element_required
            )
        VALUES
            ($1, $2, $3, $4, $5, $6)
        RETURNING
            element_id;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		row.FormID,
		row.ParamName,
		row.Caption,
		row.TemplateID,
		row.TemplateVersion,
		row.Required,
	).Scan(
		&row.ID,
	)

	return row.ID, err
}

// Update - comment method.
func (re *FormElementPostgres) Update(ctx context.Context, row entity.FormElement) (tagVersion uint32, err error) {
	set, err := re.sqlBuilder.Set().BuildEntity(row)
	if err != nil || set.Empty() {
		return 0, err
	}

	args := []any{
		row.ID,
		row.TagVersion,
	}

	setStr, setArgs := set.WithStartArg(len(args) + 1).ToSQL()

	sql := `
        UPDATE
            ` + module.DBTableNameSubmitFormElements + `
        SET
            tag_version = tag_version + 1,
			updated_at = NOW(),
            ` + setStr + `
        WHERE
            element_id = $1 AND tag_version = $2 AND deleted_at IS NULL
		RETURNING
			tag_version;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		mrsql.MergeArgs(args, setArgs)...,
	).Scan(
		&tagVersion,
	)

	return tagVersion, err
}

// Delete - comment method.
func (re *FormElementPostgres) Delete(ctx context.Context, rowID uint64) error {
	return re.repoSoftDeleter.Delete(ctx, rowID)
}
