package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Alexandrfield/Gomarket/internal/common"
)

var ErrPasswordNotValidForUser = errors.New("for this user password not valids")
var ErrOrderLoadedAnotherUser = errors.New("this num order  was used another user")
var ErrOrderLoaded = errors.New("for this num is already load")
var ErrorInsufficientFunds = errors.New("not enough points for this actions")

type StorageCommunicator interface {
	CreateNewUser(login string, password string) (string, error)
	AytorizationUser(login string, password string) (string, error) // return user_id
	IsUserLoginExist(login string) (bool, error)
	SetOrder(ord *common.UserOrder) error
	GetOrder(orderNum string) (*common.UserOrder, error)
	GetCountMarketPoints(user string) (float64, float64, error)
	UseMarketPoints(userID string, withdrawOrd *common.WithdrawOrder) error
	GetAllUserOrders(userID string) ([]common.PaymentOrder, error)
	GetAllWithdrawls(userID string) ([]common.WithdrawOrder, error)
	UpdateUserOrder(ord *common.UserOrder) error
}

func GetStorage(config Config, logger common.Logger) (StorageCommunicator, error) {
	t := DatabaseStorage{Logger: logger}
	err := t.Start(config.DatabasURI)
	if err != nil {
		return nil, fmt.Errorf("problem start DB. err:%w", err)
	}
	logger.Infof("teest Ping database: %v", t.PingDatabase())
	return &t, nil
}

type DatabaseStorage struct {
	Logger common.Logger
	db     *sql.DB
}

func (st *DatabaseStorage) createTable(ctx context.Context) error {
	st.Logger.Debugf("create table: Users")
	const queryUsers = `CREATE TABLE if NOT EXISTS Users (id SERIAL PRIMARY KEY, 
	login text, passwd text, allPoints double precision, usedPoints double precision)`
	if _, err := st.db.ExecContext(ctx, queryUsers); err != nil {
		return fmt.Errorf("error while trying to create table Users: %w", err)
	}
	st.Logger.Debugf("create table: Orders")
	const queryOrders = `CREATE TABLE if NOT EXISTS Orders (id SERIAL PRIMARY KEY, 
	numer bigint, polsak int, status text, points double precision, upload timestamp)`
	if _, err := st.db.ExecContext(ctx, queryOrders); err != nil {
		return fmt.Errorf("error while trying to create table Orders: %w", err)
	}
	st.Logger.Debugf("create table: Used")
	const queryUsed = `CREATE TABLE if NOT EXISTS Used (id SERIAL PRIMARY KEY, 
	numer bigint, polsak int, sum double precision, upload timestamp)`
	if _, err := st.db.ExecContext(ctx, queryUsed); err != nil {
		return fmt.Errorf("error while trying to create table Used: %w", err)
	}
	return nil
}
func (st *DatabaseStorage) Start(databaseDsn string) error {
	var err error
	st.db, err = sql.Open("pgx", databaseDsn)
	if err != nil {
		return fmt.Errorf("can not open database. err:%w", err)
	}
	st.Logger.Infof("Connect to db open")
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	err = st.createTable(ctx)
	if err != nil {
		errClose := st.db.Close()
		if errClose != nil {
			return fmt.Errorf("can not create table err:%w; end close connection to database err:%w",
				err, errClose)
		}
		return fmt.Errorf("can not create table. err:%w", err)
	}
	return nil
}
func (st *DatabaseStorage) PingDatabase() bool {
	status := false
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := st.db.PingContext(ctx); err == nil {
		status = true
		st.Logger.Infof("succesful ping")
	} else {
		st.Logger.Infof("Can not ping database. err:%w", err)
	}
	return status
}

func (st *DatabaseStorage) CreateNewUser(login string, password string) (string, error) {
	// check if user has been created  in handler
	tx, err := st.db.Begin()
	if err != nil {
		return "", fmt.Errorf("can not create transaction. err:%w", err)
	}
	st.Logger.Debugf("CreateNewUser login:%s;", login)

	query := `INSERT INTO Users (login, passwd, allPoints, usedPoints) VALUES ($1, $2, $3, $4)`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := tx.ExecContext(ctx, query, login, password, 0.0, 0.0); err != nil {
		errRol := tx.Rollback()
		if errRol != nil {
			return "", fmt.Errorf("error create new user login:%s. err:%w; and error rollback err:%w",
				login, err, errRol)
		}
		return "", fmt.Errorf("error create new user. login:%s: %w", login, err)
	}
	row := tx.QueryRowContext(context.Background(),
		"SELECT id FROM Users WHERE login = $1", login)
	var userID int
	err = row.Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("error scan value from row. err:%w", err)
	}
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("error with commit transactiom CreateNewUser. err:%w", err)
	}
	return strconv.Itoa(userID), nil
}
func (st *DatabaseStorage) AytorizationUser(login string, password string) (string, error) { // return user_id
	row := st.db.QueryRowContext(context.Background(),
		"SELECT id, passwd FROM Users WHERE login = $1", login)
	var userID int
	var passwd string
	err := row.Scan(&userID, &passwd)
	if err != nil {
		return "", fmt.Errorf("error scan value from row. err:%w", err)
	}
	if password != passwd {
		return "", ErrPasswordNotValidForUser
	}
	return strconv.Itoa(userID), nil
}
func (st *DatabaseStorage) IsUserLoginExist(login string) (bool, error) {
	row := st.db.QueryRowContext(context.Background(),
		"SELECT id FROM Users WHERE login = $1", login)
	var userID int
	err := row.Scan(&userID)
	if err != nil {
		return false, nil
		//return false, fmt.Errorf("error scan value from row. err:%w", err)
	}
	return true, nil
}
func (st *DatabaseStorage) SetOrder(ord *common.UserOrder) error {
	existOrd, err := st.GetOrder(ord.Ord.Number)
	if err == nil {
		if existOrd.IDUser != "" {
			if existOrd.IDUser != ord.IDUser {
				return ErrOrderLoadedAnotherUser
			} else {
				return ErrOrderLoaded
			}
		} else {
			return errors.New("not valid data. existOrd")
		}
	}
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("can not create transaction SetOrder. err:%w", err)
	}
	st.Logger.Debugf("SetOrder ord:%s;", ord)
	query := `INSERT INTO Orders (numer, polsak, status, points, upload) VALUES ($1, $2, $3, $4, $5)`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := tx.ExecContext(ctx, query, ord.Ord.Number, ord.IDUser,
		ord.Ord.Status, ord.Ord.Accural, ord.Ord.UploadedAt); err != nil {
		return fmt.Errorf("tx, error while trying to ord. err: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error with commit transactiom AddCounter. err:%w", err)
	}
	return nil
}

func (st *DatabaseStorage) GetOrder(orderNum string) (*common.UserOrder, error) {
	var userOrd common.UserOrder
	row := st.db.QueryRowContext(context.Background(),
		"SELECT  numer, polsak, status, points, upload FROM Orders WHERE numer = $1", orderNum)
	err := row.Scan(&userOrd.Ord.Number, &userOrd.IDUser, &userOrd.Ord.Status,
		&userOrd.Ord.Accural, &userOrd.Ord.UploadedAt)
	if err != nil {
		return &userOrd, fmt.Errorf("error scan value from row. err:%w", err)
	}
	return &userOrd, nil
}

func (st *DatabaseStorage) GetCountMarketPoints(userID string) (float64, float64, error) {
	row := st.db.QueryRowContext(context.Background(),
		"SELECT allPoints, usedPoints FROM Users WHERE id = $1", userID)
	fmt.Printf(">>>%s\n", row)
	var allPoints float64
	var usedPoints float64
	err := row.Scan(&allPoints, &usedPoints)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("error scan value from row. err:%w", err)
	}
	fmt.Printf(">>> allPoints:%s; usedPoints:%s\n", allPoints, usedPoints)
	return allPoints, usedPoints, nil
}
func (st *DatabaseStorage) UseMarketPoints(userID string, withdrawOrd *common.WithdrawOrder) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("can not create transaction UseMarketPoints. err:%w", err)
	}
	st.Logger.Debugf("UseMarketPointsnmae, userID:%s; value:%d;", userID, withdrawOrd)

	row := tx.QueryRowContext(context.Background(),
		"SELECT allPoints FROM Users WHERE id = $1", userID)
	var allPoints float64
	err = row.Scan(&allPoints)
	if err != nil {
		return fmt.Errorf("UseMarketPoints. error scan value from row. err:%w", err)
	}

	if allPoints < withdrawOrd.Sum {
		errr := tx.Rollback()
		if errr != nil {
			return fmt.Errorf("error UseMarketPoints: err%w;And can not rollback! err:%w",
				ErrorInsufficientFunds, errr)
		}
		return ErrorInsufficientFunds
	}

	query :=
		`UPDATE Users SET allPoints = Users.allPoints - $1, usedPoints= Users.usedPoints + $1 WHERE id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err = tx.ExecContext(ctx, query, withdrawOrd.Sum, userID); err != nil {
		errr := tx.Rollback()
		if errr != nil {
			return fmt.Errorf("error UseMarketPoints: err%w;And can not rollback! err:%w",
				ErrorInsufficientFunds, errr)
		}
		return fmt.Errorf("tx, error while trying update used points: %w", err)
	}

	query = `INSERT INTO Used (numer, polsak, sum, upload) VALUES ($1, $2, $3, $4)`
	ctxUsed, cancelUsed := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelUsed()
	if _, err = tx.ExecContext(ctxUsed, query, withdrawOrd.Order, userID,
		withdrawOrd.Sum, withdrawOrd.Processed); err != nil {
		errr := tx.Rollback()
		if errr != nil {
			return fmt.Errorf("error UseMarketPoints: err%w;And can not rollback! err:%w",
				ErrorInsufficientFunds, errr)
		}
		return fmt.Errorf("tx, error while trying update used points: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error with commit transactiom AddCounter. err:%w", err)
	}
	return nil
}

func (st *DatabaseStorage) UpdateUserOrder(ord *common.UserOrder) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("can not create transaction UpdateUserOrder. err:%w", err)
	}

	query := `UPDATE Orders SET points = $1, status = $2 WHERE id = $3`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := tx.ExecContext(ctx, query, ord.Ord.Accural, ord.Ord.Status, ord.IDUser); err != nil {
		errRol := tx.Rollback()
		if errRol != nil {
			return fmt.Errorf("error UpdateUserOrder  (update order)err:%w; and error rollback err:%w", err, errRol)
		}
		return fmt.Errorf("error 1UpdateUserOrder. err:%w", err)
	}
	queryUsers := `UPDATE Users SET allPoints = Users.allPoints + $1 WHERE id = $2`
	ctxUsers, cancelUsers := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelUsers()
	if _, err := tx.ExecContext(ctxUsers, queryUsers, ord.Ord.Accural, ord.IDUser); err != nil {
		errRol := tx.Rollback()
		if errRol != nil {
			return fmt.Errorf("error UpdateUserOrder (update user) err:%w; and error rollback err:%w", err, errRol)
		}
		return fmt.Errorf("error 2UpdateUserOrder. err:%w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error with commit transactiom UpdateUserOrder. err:%w", err)
	}
	return nil
}

func (st *DatabaseStorage) GetAllUserOrders(userID string) ([]common.PaymentOrder, error) {
	var res []common.PaymentOrder
	rows, err := st.db.QueryContext(context.Background(),
		"SELECT numer, status, points, upload FROM Orders WHERE polsak = $1", userID)
	if err != nil {
		return res, fmt.Errorf("problem GetAllUserOrders. err:%w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		userOrd := common.PaymentOrder{}
		err := rows.Scan(&userOrd.Number, &userOrd.Status, &userOrd.Accural, &userOrd.UploadedAt)
		if err != nil {
			st.Logger.Warnf("error scan value from row. err:%s", err)
		}
		res = append(res, userOrd)
	}
	err = rows.Err()
	if err != nil {
		st.Logger.Warnf("error rows. err:%s", err)
	}
	return res, nil
}
func (st *DatabaseStorage) GetAllWithdrawls(userID string) ([]common.WithdrawOrder, error) {
	var res []common.WithdrawOrder
	rows, err := st.db.QueryContext(context.Background(),
		"SELECT numer, sum, upload FROM Used WHERE polsak = $1", userID)
	if err != nil {
		return res, fmt.Errorf("problem GetAllUserOrders. err:%w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		wird := common.WithdrawOrder{}
		err := rows.Scan(&wird.Order, &wird.Sum, &wird.Processed)
		if err != nil {
			st.Logger.Warnf("error scan value from row. err:%s", err)
		}
		res = append(res, wird)
	}
	err = rows.Err()
	if err != nil {
		st.Logger.Warnf("error rows. err:%s", err)
	}
	return res, nil
}
