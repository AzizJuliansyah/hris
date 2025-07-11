package models

import (
	"database/sql"
	"hris/config"
	"hris/entities"
	"log"
	"time"

	"github.com/goodsign/monday"
)

type NewsModel struct {
	db *sql.DB
}

func NewNewsModel() *NewsModel {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Failed connect to database:", err)
	}
	return &NewsModel{
		db: conn,
	}
}

func (model NewsModel) AddNews(news entities.News) error {
	_, err := model.db.Exec(
		"INSERT INTO news (creator_nik, assigne_nik, thumbnail, title, content, footer, start_date, end_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		news.Creator_NIK, news.Assigne_NIK, news.Thumbnail, news.Title, news.Content, news.Footer, news.Start_Date, news.End_Date,
	)
	return err
}

func (model NewsModel) FindAllNews() ([]entities.News, error) {
	var created_atTime time.Time
	var thumbnail sql.NullString
	var creatorName sql.NullString
	var assigneName sql.NullString

	rows, err := model.db.Query(`
		SELECT 
			news.id,
			news.assigne_nik,
			news.thumbnail,
			news.title,
			news.content,
			news.footer,
			news.start_date,
			news.end_date,
			news.created_at,
			creator.name AS creator_name,
			assigne.name AS assigne_name
		FROM news
		JOIN employee AS creator ON news.creator_nik = creator.nik
		LEFT JOIN employee AS assigne ON news.assigne_nik = assigne.nik
		WHERE news.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newss []entities.News
	for rows.Next() {
		var news entities.News
		err := rows.Scan(
			&news.Id,
			&news.Assigne_NIK,
			&thumbnail,
			&news.Title,
			&news.Content,
			&news.Footer,
			&news.Start_Date,
			&news.End_Date,
			&created_atTime,
			&creatorName,
			&assigneName,
		)
		if err != nil {
			return nil, err
		}
		if creatorName.Valid {
			news.Creator_Name = creatorName.String
		} else {
			news.Creator_Name = "undefined"
		}
		if assigneName.Valid {
			news.Assigne_Name = assigneName.String
		} else {
			news.Assigne_Name = "-"
		}

		news.Created_at = created_atTime
		news.Created_atFormat = monday.Format(created_atTime, "02 Januari 2006 15:04", monday.LocaleIdID)
		
		if thumbnail.Valid {
			news.Thumbnail = thumbnail
		} else {
			news.Thumbnail = sql.NullString{String: "", Valid: false}
		}
		
		newss = append(newss, news)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return newss, nil
}


func (model NewsModel) CountAllNews() (int, error) {
	row := model.db.QueryRow("SELECT COUNT(*) FROM news WHERE deleted_at IS NULL")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (model NewsModel) FindNewsByID(id int64) (entities.News, error) {
	var news entities.News
	var assigne_NIK sql.NullString
	query := `
		SELECT id, assigne_nik, thumbnail, title, content, footer, start_date, end_date, created_at
		FROM news 
		WHERE id = ? AND deleted_at IS NULL
	`
	err := model.db.QueryRow(query, id).Scan(
		&news.Id,
		&assigne_NIK,
		&news.Thumbnail,
		&news.Title,
		&news.Content,
		&news.Footer,
		&news.Start_Date,
		&news.End_Date,
		&news.Created_at,
	)
	if assigne_NIK.Valid {
		news.Assigne_NIK = assigne_NIK
	} else {
		news.Assigne_NIK = sql.NullString{String: "", Valid: false}
	}
	
	return news, err
}

func (model NewsModel) GetThumbnailByID(id int64) (sql.NullString, error) {
	var thumbnail sql.NullString
	err := model.db.QueryRow("SELECT thumbnail FROM news WHERE id = ? AND deleted_at IS NULL", id).Scan(&thumbnail)
	return thumbnail, err
}

func (model NewsModel) EditNews(news entities.News) error {
	query := `
		UPDATE news 
		SET updated_creator_nik = ?, assigne_nik = ?, thumbnail = ?, title = ?, content = ?, footer = ?, start_date = ?, end_date = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(
		query,
		news.Updated_Creator_NIK,
		news.Assigne_NIK,
		news.Thumbnail,
		news.Title,
		news.Content,
		news.Footer,
		news.Start_Date,
		news.End_Date,
		time.Now(),
		news.Id,
	)
	return err
}

func (model NewsModel) SoftDeleteNews(id int64) error {
	query := `
		UPDATE news 
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := model.db.Exec(query, time.Now(), id)
	return err
}