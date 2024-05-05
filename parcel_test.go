package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убедждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	
	require.NoError(t, err)
	parcel.Number = id
	require.NotEmpty(t, id)
	par, err := store.Get(id)
	// get
	// получаем только что добавленную посылку, убеждаемся в отсутствии ошибки
	require.NoError(t, err)
	// проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	require.Equal(t, parcel, par)


	// delete
	// удаляем добавленную посылку, убеждаемся в отсутствии ошибки
	
	err = store.Delete(id)
	require.NoError(t, err)
// проверяем, что посылку больше нельзя получить из БД

	_, err = store.Get(id)
    require.Equal(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	
	require.NoError(t, err)
	require.NotEmpty(t, id)
	// set address
	// обновляем адрес, убеждаемся в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)

	require.NoError(t, err)
	// check
	// получаем добавленную посылку и убеждаемся, что адрес обновился
	par, err := store.Get(id)
	assert.Equal(t, newAddress, par.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	
	require.NoError(t, err)
	require.NotEmpty(t, id)
	// set status
	// обновляем статус, убеждаемся в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)

	require.NoError(t, err)
	// check
	// получаем добавленную посылку и убеждаемся, что статус обновился
	par, err := store.Get(id)
	assert.Equal(t, ParcelStatusSent, par.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcel)
	
		require.NoError(t, err)
		require.NotEmpty(t, id)// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)// получаем список посылок по идентификатору клиента, сохранённого в переменной client
	// убеждаемся в отсутствии ошибки
	require.NoError(t, err)
	// убеждаемся, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(parcels), len(parcelMap))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убеждаемся, что все посылки из storedParcels есть в parcelMap
		// убеждаемся, что значения полей полученных посылок заполнены верно
		ParcelFromMap, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		require.Equal(t, ParcelFromMap, parcel)
	}
}
