package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/usememos/memos/plugin/filter"
	"github.com/usememos/memos/store"
)

func (d *DB) UpsertMemoRelation(ctx context.Context, create *store.MemoRelation) (*store.MemoRelation, error) {
	stmt := `
		INSERT INTO memo_relation (
			memo_id,
			related_memo_id,
			type
		)
		VALUES (` + placeholders(3) + `)
		RETURNING memo_id, related_memo_id, type
	`
	memoRelation := &store.MemoRelation{}
	if err := d.db.QueryRowContext(
		ctx,
		stmt,
		create.MemoID,
		create.RelatedMemoID,
		create.Type,
	).Scan(
		&memoRelation.MemoID,
		&memoRelation.RelatedMemoID,
		&memoRelation.Type,
	); err != nil {
		return nil, err
	}

	return memoRelation, nil
}

func (d *DB) ListMemoRelations(ctx context.Context, find *store.FindMemoRelation) ([]*store.MemoRelation, error) {
	where, args := []string{"1 = 1"}, []any{}
	if find.MemoID != nil {
		where, args = append(where, "memo_id = "+placeholder(len(args)+1)), append(args, find.MemoID)
	}
	if find.RelatedMemoID != nil {
		where, args = append(where, "related_memo_id = "+placeholder(len(args)+1)), append(args, find.RelatedMemoID)
	}
	if find.Type != nil {
		where, args = append(where, "type = "+placeholder(len(args)+1)), append(args, find.Type)
	}
	if find.MemoFilter != nil {
		// Parse filter string and return the parsed expression.
		// The filter string should be a CEL expression.
		parsedExpr, err := filter.Parse(*find.MemoFilter, filter.MemoFilterCELAttributes...)
		if err != nil {
			return nil, err
		}
		convertCtx := filter.NewConvertContext()
		convertCtx.ArgsOffset = len(args)
		// ConvertExprToSQL converts the parsed expression to a SQL condition string.
		converter := filter.NewCommonSQLConverterWithOffset(&filter.PostgreSQLDialect{}, convertCtx.ArgsOffset+len(convertCtx.Args))
		if err := converter.ConvertExprToSQL(convertCtx, parsedExpr.GetExpr()); err != nil {
			return nil, err
		}
		condition := convertCtx.Buffer.String()
		if condition != "" {
			where = append(where, fmt.Sprintf("memo_id IN (SELECT id FROM memo WHERE %s)", condition))
			where = append(where, fmt.Sprintf("related_memo_id IN (SELECT id FROM memo WHERE %s)", condition))
			args = append(args, convertCtx.Args...)
		}
	}

	rows, err := d.db.QueryContext(ctx, `
		SELECT
			memo_id,
			related_memo_id,
			type
		FROM memo_relation
		WHERE `+strings.Join(where, " AND "), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*store.MemoRelation{}
	for rows.Next() {
		memoRelation := &store.MemoRelation{}
		if err := rows.Scan(
			&memoRelation.MemoID,
			&memoRelation.RelatedMemoID,
			&memoRelation.Type,
		); err != nil {
			return nil, err
		}
		list = append(list, memoRelation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (d *DB) DeleteMemoRelation(ctx context.Context, delete *store.DeleteMemoRelation) error {
	where, args := []string{"1 = 1"}, []any{}
	if delete.MemoID != nil {
		where, args = append(where, "memo_id = "+placeholder(len(args)+1)), append(args, delete.MemoID)
	}
	if delete.RelatedMemoID != nil {
		where, args = append(where, "related_memo_id = "+placeholder(len(args)+1)), append(args, delete.RelatedMemoID)
	}
	if delete.Type != nil {
		where, args = append(where, "type = "+placeholder(len(args)+1)), append(args, delete.Type)
	}
	stmt := `DELETE FROM memo_relation WHERE ` + strings.Join(where, " AND ")
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}
	if _, err = result.RowsAffected(); err != nil {
		return err
	}
	return nil
}
