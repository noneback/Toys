package db

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

const (
	NotFinished = "NotFinished"
	Finished    = "Finished"
)

func init() {
	var err error
	TaskBucket = []byte("task")
	CompleteBucket = []byte("isComplete")
	DB, err = bolt.Open("task.db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}

	DB.Update(CreateBucket(string(TaskBucket)))
	DB.Update(CreateBucket(string(CompleteBucket)))

	//defer DB.Close()

}

var TaskBucket []byte
var CompleteBucket []byte

type action func(tx *bolt.Tx) error

func AddTask(taskDes string) action {
	return func(tx *bolt.Tx) error {
		tb := tx.Bucket(TaskBucket)
		seq, _ := tb.NextSequence()
		id := int(seq)

		if err := tb.Put(itob(id), []byte(taskDes)); err != nil {
			return nil
		}

		cb := tx.Bucket(CompleteBucket)
		if err := cb.Put(itob(id), []byte(NotFinished)); err != nil {
			return err
		}

		return nil

	}
}

func CreateBucket(bucketName string) action {
	return func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}
}

func ShowList(flag string) action {
	return func(tx *bolt.Tx) error {
		tb := tx.Bucket(TaskBucket)
		c := tb.Cursor()

		cb := tx.Bucket(CompleteBucket)
		for k, v := c.First(); k != nil; k, v = c.Next() {

			stat := cb.Get(k)

			if flag == "all" {
				fmt.Printf("%s.%s\t| %s\n", k, v, stat)
			} else if strings.ToLower(string(stat)) == flag {
				fmt.Printf("%s.%s\t| %s\n", k, v, stat)
			}
		}
		return nil
	}
}

func FinishTask(id int) action {
	return func(tx *bolt.Tx) error {
		cb := tx.Bucket(CompleteBucket)
		if err := cb.Put(itob(id), []byte(Finished)); err != nil {
			return fmt.Errorf("no such a task")
		}
		return nil

	}
}

func RemoveTask(id int) action {
	return func(tx *bolt.Tx) error {
		tb := tx.Bucket(TaskBucket)
		cb := tx.Bucket(CompleteBucket)
		if err := tb.Delete(itob(id)); err != nil {
			return fmt.Errorf("no such a task")
		}
		if err := cb.Delete(itob(id)); err != nil {
			return fmt.Errorf("no such a task")
		}
		return nil
	}
}

func itob(v int) []byte {
	return []byte(strconv.Itoa(v))
}

var DB *bolt.DB
