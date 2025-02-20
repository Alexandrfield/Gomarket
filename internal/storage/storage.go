package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Alexandrfield/Gomarket/internal/common"
)

var ErrPasswordNotValidForUser = errors.New("For this user password not valids")
var ErrOrderLoadedAnotherUser = errors.New("This num order  was used another user")
var ErrOrderLoaded = errors.New("For this num is already load")
var ErrorInsufficientFunds = errors.New("Not enough points for this actions")

type StorageCommunicator interface {
	CreateNewUser(login string, password string) (string, error)
	AytorizationUser(login string, password string) (string, error) // return user_id
	IsUserLoginExist(login string) (bool, error)
	SetOrder(ord common.UserOrder) error
	GetOrder(orderNum string) (common.UserOrder, error)
	GetCountMarketPoints(user string) (float64, float64, error)
	UseMarketPoints(userId string, withdrawOrd common.WithdrawOrder) error
	GetAllUserOrders(userId string) ([]common.PaymentOrder, error)
	GetAllWithdrawls(userId string) ([]common.WithdrawOrder, error)
	UpdateUserOrder(ord common.UserOrder) error
}

func GetStorage(config Config, logger common.Logger) (StorageCommunicator, error) {
	t := DatabaseStorage{Logger: logger}
	fmt.Printf("teest 1 \n")
	err := t.Start(config.DatabasURI)
	if err != nil {
		return nil, fmt.Errorf("problem start DB. err:%w", err)
	}
	fmt.Printf("teest 2 \n")
	fmt.Printf("teest Ping ->>> %v", t.PingDatabase())
	return &t, nil
}

type DatabaseStorage struct {
	Logger common.Logger
	db     *sql.DB
}

func (st *DatabaseStorage) createTable(ctx context.Context) error {
	const queryUsers = `CREATE TABLE if NOT EXISTS Users (id int PRIMARY KEY, 
	login text, passwd text, allPoints double precision, usedPoints double precision)`
	if _, err := st.db.ExecContext(ctx, queryUsers); err != nil {
		return fmt.Errorf("error while trying to create table Users: %w", err)
	}
	const queryOrders = `CREATE TABLE if NOT EXISTS Orders (id int PRIMARY KEY, 
	numer bigint, user int, status text, points double precision, upload timestamp)`
	if _, err := st.db.ExecContext(ctx, queryOrders); err != nil {
		return fmt.Errorf("error while trying to create table Orders: %w", err)
	}
	const queryUsed = `CREATE TABLE if NOT EXISTS Used (id int PRIMARY KEY, 
	numer bigint, sum double precision, upload timestamp)`
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
			return fmt.Errorf("can not create table err:%w; end close connection to database err:%w", err, errClose)
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
	st.Logger.Debugf("CreateNewUserю login:%s;", login)

	query := `INSERT INTO User (login, passwd, allPoints, usedPoints) VALUES ($1, $2, $3, $4)`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := tx.ExecContext(ctx, query, login, password, 0.0, 0.0); err != nil {
		errRol := tx.Rollback()
		if errRol != nil {
			return "", fmt.Errorf("error create new user login:%s. err:%w; and error rollback err:%w", login, err, errRol)
		}
		return "", fmt.Errorf("error create new user. login:%s: %w", login, err)
	}
	row := tx.QueryRowContext(context.Background(),
		"SELECT id FROM Users WHERE login = $1", login)
	var userId int
	err = row.Scan(&userId)
	if err != nil {
		return "", fmt.Errorf("error scan value from row. err:%w", err)
	}
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("error with commit transactiom CreateNewUser. err:%w", err)
	}
	return strconv.Itoa(userId), nil
}
func (st *DatabaseStorage) AytorizationUser(login string, password string) (string, error) { // return user_id

	row := st.db.QueryRowContext(context.Background(),
		"SELECT id, passwd FROM Users WHERE login = $1", login)
	var userId int
	var passwd string
	err := row.Scan(&userId, &passwd)
	if err != nil {
		return "", fmt.Errorf("error scan value from row. err:%w", err)
	}
	if password != passwd {
		return "", ErrPasswordNotValidForUser
	}

	return strconv.Itoa(userId), nil
}
func (st *DatabaseStorage) IsUserLoginExist(login string) (bool, error) {
	row := st.db.QueryRowContext(context.Background(),
		"SELECT id FROM Users WHERE login = $1", login)
	var userId int
	err := row.Scan(&userId)
	if err != nil {
		return false, fmt.Errorf("error scan value from row. err:%w", err)
	}
	return true, nil
}
func (st *DatabaseStorage) SetOrder(ord common.UserOrder) error {

	existOrd, err := st.GetOrder(ord.Ord.Number)
	if err == nil {
		if existOrd.IDUser != "" {
			if existOrd.IDUser != ord.IDUser {
				return ErrOrderLoadedAnotherUser
			} else {
				return ErrOrderLoaded
			}
		} else {
			return fmt.Errorf("Not valid data. %s", existOrd)
		}
	}
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("can not create transaction SetOrder. err:%w", err)
	}
	st.Logger.Debugf("SetOrder ord:%s;", ord)
	query := `INSERT INTO Orders (numer, user, status, points, upload) VALUES ($1, $2, $3, $4, %5)`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := tx.ExecContext(ctx, query, ord.Ord.Number, ord.IDUser, ord.Ord.Status, ord.Ord.Accural, ord.Ord.Uploaded_at); err != nil {
		return fmt.Errorf("tx, error while trying to ordc %s: %w", ord, err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error with commit transactiom AddCounter. err:%w", err)
	}

	return nil
}

func (st *DatabaseStorage) GetOrder(orderNum string) (common.UserOrder, error) {
	var userOrd common.UserOrder
	row := st.db.QueryRowContext(context.Background(),
		"SELECT  numer, user, status, points, upload FROM Orders WHERE numer = $1", orderNum)
	err := row.Scan(&userOrd.Ord.Number, &userOrd.IDUser, &userOrd.Ord.Status, &userOrd.Ord.Accural, &userOrd.Ord.Uploaded_at)
	if err != nil {
		return userOrd, fmt.Errorf("error scan value from row. err:%w", err)
	}
	return userOrd, nil
}

//	func (st *DatabaseStorage) GetOrderStatus(user string, order string) (string, error) {
//		return "test", nil
//	}
func (st *DatabaseStorage) GetCountMarketPoints(userId string) (float64, float64, error) {
	row := st.db.QueryRowContext(context.Background(),
		"SELECT allPoints, usedPoints FROM Users WHERE id = $1", userId)
	var allPoints float64
	var usedPoints float64
	err := row.Scan(&allPoints, &usedPoints)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("error scan value from row. err:%w", err)
	}

	return allPoints, usedPoints, nil
}
func (st *DatabaseStorage) UseMarketPoints(userId string, withdrawOrd common.WithdrawOrder) error {
	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("can not create transaction UseMarketPoints. err:%w", err)
	}
	st.Logger.Debugf("UseMarketPointsnmae, userId:%s; value:%d;", userId, withdrawOrd)

	row := tx.QueryRowContext(context.Background(),
		"SELECT allPoints FROM Users WHERE id = $1", userId)
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

	query := `UPDATE Orders SET  allPoints = Orders.allPoints - $1, usedPoints= Orders.usedPoints + $1 WHERE id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err = tx.ExecContext(ctx, query, withdrawOrd.Sum, userId); err != nil {
		errr := tx.Rollback()
		if errr != nil {
			return fmt.Errorf("error UseMarketPoints: err%w;And can not rollback! err:%w",
				ErrorInsufficientFunds, errr)
		}
		return fmt.Errorf("tx, error while trying update used points: %w", err)
	}

	query = `INSERT INTO Used (numer, sum, upload) VALUES ($1, $2, $3)`
	ctxUsed, cancelUsed := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelUsed()
	if _, err = tx.ExecContext(ctxUsed, query, withdrawOrd.Order, withdrawOrd.Sum, withdrawOrd.Processed); err != nil {
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

func (st *DatabaseStorage) UpdateUserOrder(ord common.UserOrder) error {
	query := `UPDATE Orders SET points = $1, status = $2 WHERE id = $3`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := st.db.ExecContext(ctx, query, ord.Ord.Accural, ord.Ord.Status, ord.IDUser); err != nil {
		return fmt.Errorf("UpdateUserOrder, error while trying update used points: %w", err)
	}

	return nil
}

func (st *DatabaseStorage) GetAllUserOrders(userId string) ([]common.PaymentOrder, error) {
	var res []common.PaymentOrder
	rows, err := st.db.QueryContext(context.Background(),
		"SELECT numer, status, points, upload FROM Orders WHERE user = $1", userId)
	if err != nil {
		return res, fmt.Errorf("Problem GetAllUserOrders. err:%w", err)
	}
	for rows.Next() {
		userOrd := common.PaymentOrder{}
		err := rows.Scan(&userOrd.Number, &userOrd.Status, &userOrd.Accural, &userOrd.Uploaded_at)
		if err != nil {
			st.Logger.Warnf("error scan value from row. err:%s", err)
		}
		res = append(res, userOrd)
	}
	return res, nil
}
func (st *DatabaseStorage) GetAllWithdrawls(userId string) ([]common.WithdrawOrder, error) {
	var res []common.WithdrawOrder
	rows, err := st.db.QueryContext(context.Background(),
		"SELECT numer, sum, upload FROM Used WHERE user = $1", userId)
	if err != nil {
		return res, fmt.Errorf("Problem GetAllUserOrders. err:%w", err)
	}
	for rows.Next() {
		wird := common.WithdrawOrder{}
		err := rows.Scan(&wird.Order, &wird.Sum, &wird.Processed)
		if err != nil {
			st.Logger.Warnf("error scan value from row. err:%s", err)
		}
		res = append(res, wird)
	}
	return res, nil

}
