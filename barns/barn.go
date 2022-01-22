package barns

import (
	"database/sql"
	"errors"
)

type Barn struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Owner struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	UserID string `json:"user_id"`
}

type sqlOwner struct {
	Owner
	Name sql.NullString
}

type BarnOwner struct {
	ID      int64 `json:"id"`
	BarnID  int64 `json:"barn_id"`
	OwnerID int64 `json:"owner_id"`
}

func (b *Barn) Save(userID string, db *sql.DB) error {
	query := "insert into barns (name) values (?)"
	result, err := db.Exec(query, b.Name)
	if err != nil {
		return errors.New("failed to insert barn into database: " + err.Error())
	}
	b.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	owner, err := HandleOwner(userID, db)
	if err != nil {
		return errors.New("failed to handle owner: " + err.Error())
	}

	// add user as owner of barn
	_, err = NewBarnOwner(b.ID, owner.ID, db)
	if err != nil {
		return errors.New("failed to add barn to owner: " + err.Error())
	}
	return nil
}

func GetBarnsByUserID(userID string, db *sql.DB) ([]*Barn, error) {
	query := "select b.id, b.name from barns b join barn_owners bo on b.id = bo.barn_id join owners o on bo.owner_id = o.id where o.user_id = ?"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, errors.New("failed to select barns from database: " + err.Error())
	}
	defer rows.Close()
	var barns []*Barn
	for rows.Next() {
		var b Barn
		err := rows.Scan(&b.ID, &b.Name)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		barns = append(barns, &b)
	}
	return barns, nil
}

func HandleOwner(userID string, db *sql.DB) (*Owner, error) {
	var owner Owner
	owner.UserID = userID
	query := "select id, name from owners where user_id = ?"
	row := db.QueryRow(query, userID)
	var o sqlOwner
	err := row.Scan(&o.ID, &o.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			// create owner
			insert := "insert into owners (user_id) values (?)"
			result, err := db.Exec(insert, owner.UserID)
			if err != nil {
				return nil, errors.New("failed to insert owner into database: " + err.Error())
			}
			owner.ID, err = result.LastInsertId()
			if err != nil {
				return nil, errors.New("failed to get last insert ID: " + err.Error())
			}
		} else {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
	}
	owner.ID = o.ID
	owner.Name = o.Name.String
	return &owner, nil
}

func NewBarnOwner(barnID int64, ownerID int64, db *sql.DB) (*BarnOwner, error) {
	query := "insert into barn_owners (barn_id, owner_id) values (?, ?)"
	result, err := db.Exec(query, barnID, ownerID)
	if err != nil {
		return nil, errors.New("failed to insert barn owner into database: " + err.Error())
	}
	var barnOwner BarnOwner
	barnOwner.ID, err = result.LastInsertId()
	if err != nil {
		return nil, errors.New("failed to get last insert ID: " + err.Error())
	}
	barnOwner.BarnID = barnID
	barnOwner.OwnerID = ownerID
	return &barnOwner, nil
}
