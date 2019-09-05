package scheduler

import (
	"CronServerMbs/database"
	"CronServerMbs/functions"
	"fmt"
	"github.com/leekchan/accounting"
	"github.com/parnurzeal/gorequest"
	_ "github.com/shopspring/decimal"
	"gopkg.in/gomail.v2"
	_ "math/big"
	"net/http"
	"strconv"
	"strings"
)

var conn = database.ConnectDB()

type SendMessageJson struct {
	cHP    string `json:"phone"`
	cPesan string `json:"message"`
}

type SelectMessage struct {
	MessageId     int    `json:"chat_id"`
	MessageNumber string ` json:"receiver_number"`
	MessageText   string `json:"message"`
	MessageStatus string `"json:"status"`
}

type SelectMessageEmail struct {
	MessageId      int    `json:"chat_id"`
	MessageEmail   string ` json:"email"`
	MessageSubject string `json:"subject"`
	MessageContent string `json:"content"`
	MessageStatus  string `"json:"status"`
}

type SelectNotifikasi struct {
	Id          string  `json:"id"`
	NoRekening  string  `json:"no_rekening"`
	NamaNasabah string  `json:"nama_nasabah"`
	Keterangan  string  `json:"keterangan"`
	Nominal     float64 `json:"nominal"`
	MyKodeTrans string  `json:"my_kode_trans"`
	TglTrans    string  `json:"tgl_trans"`
	Modul       string  `json:"modul"`
	Hp          string  `json:"hp"`
	StatusWA    string  `json:"status_wa"`
	Jam         string  `json:"jam"`
	Email       string  `json:"email"`
	StatusEMail string  `json:"status_email"`
	Subject     string  `json:"subject"`
}

func CekNotifikasiWA() {
	sqlStatement := "SELECT id," +
		"ifnull(concat(left(no_rekening,3),replace(no_rekening,left(no_rekening,length(no_rekening)-3),'XXX')),'') AS no_rekening," +
		"ifnull(nama_nasabah,'') nama_nasabah,ifnull(keterangan,'') keterangan,coalesce(nominal,0) nominal," +
		"ifnull(my_kode_trans,'') my_kode_trans,ifnull(date(tgl_trans),date(now())) tgl_trans,ifnull(modul,'') modul,ifnull(hp,'') hp,ifnull(status_wa,0) status_wa,time(jam) as jam,ifnull(email,'') email,ifnull(status_email,0) status_email,ifnull(subject,'') subject " +
		"FROM wa_notifikasi where STATUS_WA=0 and hp<>'' "
	rows, err := database.ConnectDB().Query(sqlStatement)
	if err != nil {
		functions.Logger().Error(err.Error())
		return
	}
	defer database.ConnectDB().Close()

	for rows.Next() {
		functions.Logger().Info("Starting Scheduler Cek Notif WA")
		messageList := SelectNotifikasi{}
		err = rows.Scan(&messageList.Id, &messageList.NoRekening, &messageList.NamaNasabah,
			&messageList.Keterangan, &messageList.Nominal, &messageList.MyKodeTrans,
			&messageList.TglTrans, &messageList.Modul, &messageList.Hp,
			&messageList.StatusWA, &messageList.Jam, &messageList.Email, &messageList.StatusEMail, &messageList.Subject)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}
		TextMessage := messageList.Keterangan
		Modul := messageList.Modul
		if Modul != "CUSTOM" {
			TextMessage = GetFValueByFKeyValue("template", "template_name", messageList.Modul, "template_text")
		}
		NamaLembaga := "BPR TESTING"
		NamaLembaga = GetFValueByFKeyValue("config", "config_name", "nama_lembaga", "config_value")
		AlamatLembaga := "JL RAYA"
		AlamatLembaga = GetFValueByFKeyValue("config", "config_name", "alamat_lembaga", "config_value")
		TaglineLembaga := "BERSAMA KITA MAJU"
		TaglineLembaga = GetFValueByFKeyValue("config", "config_name", "tagline_lembaga", "config_value")
		NoLembaga := "0"
		NoLembaga = GetFValueByFKeyValue("config", "config_name", "telp_lembaga", "config_value")

		ac := accounting.Accounting{Symbol: "Rp ", Precision: 2, Thousand: ".", Decimal: ","}
		TextMessage = strings.Replace(TextMessage, "[no_rekening]", messageList.NoRekening, 99)
		TextMessage = strings.Replace(TextMessage, "[nama_lembaga]", NamaLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[alamat_lembaga]", AlamatLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[tagline_lembaga]", TaglineLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[nama_nasabah]", messageList.NamaNasabah, 99)
		TextMessage = strings.Replace(TextMessage, "[tgl_trans]", messageList.TglTrans, 99)
		TextMessage = strings.Replace(TextMessage, "[jam]", messageList.Jam, 99)
		TextMessage = strings.Replace(TextMessage, "[no_telepon]", NoLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[nominal]", ac.FormatMoney(messageList.Nominal), 99)
		Id := messageList.Id
		stmt, err := conn.Prepare("UPDATE wa_notifikasi SET STATUS_WA=1 where id=?")
		if err != nil {
			//panic(err.Error())
			functions.Logger().Error(err.Error())
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(Id)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}

		stmt, err = conn.Prepare("Insert into outbox (receiver_number,message,status) values (?,?,?)")
		if err != nil {
			//panic(err.Error())
			functions.Logger().Error(err.Error())
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(messageList.Hp, TextMessage, 0)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}

		functions.Logger().Info("Successfully Cek Notif WA")
	}

	rows.Close()
}

func CekNotifikasiEmail() {
	sqlStatement := "SELECT id," +
		"ifnull(concat(left(no_rekening,3),replace(no_rekening,left(no_rekening,length(no_rekening)-3),'XXX')),'') AS no_rekening," +
		"ifnull(nama_nasabah,'') nama_nasabah,ifnull(keterangan,'') keterangan,coalesce(nominal,0) nominal," +
		"ifnull(my_kode_trans,'') my_kode_trans,ifnull(date(tgl_trans),date(now())) tgl_trans,ifnull(modul,'') modul,ifnull(hp,'') hp,ifnull(status_wa,0) status_wa,time(jam) as jam,ifnull(email,'') email,ifnull(status_email,0) status_email,ifnull(subject,'') subject " +
		"FROM wa_notifikasi " +
		"where STATUS_EMAIL=0 and email<>'' and use_email=1 "
	rows, err := database.ConnectDB().Query(sqlStatement)
	if err != nil {
		functions.Logger().Error(err.Error())
		return
	}
	defer database.ConnectDB().Close()
	for rows.Next() {
		functions.Logger().Info("Starting Scheduler Cek Notif Email")
		messageList := SelectNotifikasi{}
		err = rows.Scan(&messageList.Id, &messageList.NoRekening, &messageList.NamaNasabah,
			&messageList.Keterangan, &messageList.Nominal, &messageList.MyKodeTrans,
			&messageList.TglTrans, &messageList.Modul, &messageList.Hp,
			&messageList.StatusWA, &messageList.Jam, &messageList.Email, &messageList.StatusEMail, &messageList.Subject)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}
		TextMessage := messageList.Keterangan
		Modul := messageList.Modul
		if Modul != "CUSTOM" {
			TextMessage = GetFValueByFKeyValue("template_email", "template_name", messageList.Modul, "template_text")
		}
		NamaLembaga := "BPR TESTING"
		NamaLembaga = GetFValueByFKeyValue("config", "config_name", "nama_lembaga", "config_value")
		AlamatLembaga := "JL RAYA"
		AlamatLembaga = GetFValueByFKeyValue("config", "config_name", "alamat_lembaga", "config_value")
		TaglineLembaga := "BERSAMA KITA MAJU"
		TaglineLembaga = GetFValueByFKeyValue("config", "config_name", "tagline_lembaga", "config_value")
		NoLembaga := "0"
		NoLembaga = GetFValueByFKeyValue("config", "config_name", "telp_lembaga", "config_value")
		Keterangan := messageList.Keterangan

		ac := accounting.Accounting{Symbol: "Rp ", Precision: 2, Thousand: ".", Decimal: ","}
		TextMessage = strings.Replace(TextMessage, "[no_rekening]", messageList.NoRekening, 99)
		TextMessage = strings.Replace(TextMessage, "[nama_lembaga]", NamaLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[alamat_lembaga]", AlamatLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[tagline_lembaga]", TaglineLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[nama_nasabah]", messageList.NamaNasabah, 99)
		TextMessage = strings.Replace(TextMessage, "[tgl_trans]", messageList.TglTrans, 99)
		TextMessage = strings.Replace(TextMessage, "[jam]", messageList.Jam, 99)
		TextMessage = strings.Replace(TextMessage, "[no_telepon]", NoLembaga, 99)
		TextMessage = strings.Replace(TextMessage, "[keterangan]", Keterangan, 99)
		TextMessage = strings.Replace(TextMessage, "[nominal]", ac.FormatMoney(messageList.Nominal), 99)
		cJenisTransaksi := ""
		if messageList.Modul == "TABSETORAN" {
			cJenisTransaksi = "Setoran Tabungan"
		} else if messageList.Modul == "TABTARIK" {
			cJenisTransaksi = "Penarikan Tabungan"
		} else if messageList.Modul == "DEPSETORPOKOK" {
			cJenisTransaksi = "Setoran Pokok Deposito"
		} else if messageList.Modul == "DEPTARIKPOKOK" {
			cJenisTransaksi = "Penarikan Pokok Deposito"
		} else if messageList.Modul == "DEPSETORBUNGA" {
			cJenisTransaksi = "Setoran Bunga Deposito"
		} else if messageList.Modul == "KREREALISASI" {
			cJenisTransaksi = "Pencairan Kredit"
		} else if messageList.Modul == "KREANGSUR" {
			cJenisTransaksi = "Angsuran Kredit"
		} else if messageList.Modul == "KRETAGIHAN" {
			cJenisTransaksi = "Tagihan Kredit"
		}

		TextMessage = strings.Replace(TextMessage, "[jenis_transaksi]", cJenisTransaksi, 99)
		Subjek := messageList.Subject
		ac = accounting.Accounting{Symbol: "Rp ", Precision: 2, Thousand: ".", Decimal: ","}
		Subjek = strings.Replace(Subjek, "[no_rekening]", messageList.NoRekening, 99)
		Subjek = strings.Replace(Subjek, "[nama_lembaga]", NamaLembaga, 99)
		Subjek = strings.Replace(Subjek, "[alamat_lembaga]", AlamatLembaga, 99)
		Subjek = strings.Replace(Subjek, "[tagline_lembaga]", TaglineLembaga, 99)
		Subjek = strings.Replace(Subjek, "[nama_nasabah]", messageList.NamaNasabah, 99)
		Subjek = strings.Replace(Subjek, "[tgl_trans]", messageList.TglTrans, 99)
		Subjek = strings.Replace(Subjek, "[jam]", messageList.Jam, 99)
		Subjek = strings.Replace(Subjek, "[no_telepon]", NoLembaga, 99)
		Subjek = strings.Replace(Subjek, "[nominal]", ac.FormatMoney(messageList.Nominal), 99)
		Id := messageList.Id
		stmt, err := conn.Prepare("UPDATE wa_notifikasi SET STATUS_EMAIL=1 where id=?")
		if err != nil {
			//panic(err.Error())
			functions.Logger().Error(err.Error())
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(Id)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}

		stmt, err = conn.Prepare("Insert into outbox_email (email,content,status,subject) values (?,?,?,?)")
		if err != nil {
			//panic(err.Error())
			functions.Logger().Error(err.Error())
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(messageList.Email, TextMessage, 0, Subjek)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}

		functions.Logger().Info("Successfully Cek Notif Email")
	}
	rows.Close()
}
func CekOutboxWA() {
	sqlStatement := "SELECT chat_id,receiver_number,message,status FROM outbox where STATUS=0"
	rows, err := database.ConnectDB().Query(sqlStatement)
	if err != nil {
		functions.Logger().Error(err.Error())
		return
	}
	defer database.ConnectDB().Close()
	for rows.Next() {
		functions.Logger().Info("Starting Send Message to WABLAS")
		messageList := SelectMessage{}
		err = rows.Scan(&messageList.MessageId, &messageList.MessageNumber, &messageList.MessageText, &messageList.MessageStatus)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}
		ret := SendMessageWA(messageList.MessageNumber, messageList.MessageText, messageList.MessageId)
		if ret > 0 {
			functions.Logger().Info("Successfully Send Message to WABLAS " + strconv.Itoa(messageList.MessageId) + "")
		} else {
			functions.Logger().Info("Failed Send Message to WABLAS " + strconv.Itoa(messageList.MessageId) + "")
		}
	}
	rows.Close()
}

func CekOutboxEmail() {
	sqlStatement := "SELECT chat_id,email,subject,content,status FROM outbox_email where STATUS=0"
	rows, err := database.ConnectDB().Query(sqlStatement)
	if err != nil {
		functions.Logger().Error(err.Error())
		return
	}
	defer database.ConnectDB().Close()
	for rows.Next() {
		functions.Logger().Info("Starting Send Email to API")
		messageList := SelectMessageEmail{}
		err = rows.Scan(&messageList.MessageId, &messageList.MessageEmail, &messageList.MessageSubject,
			&messageList.MessageContent, &messageList.MessageStatus)
		if err != nil {
			functions.Logger().Error(err.Error())
			return
		}
		ret := sendMail(messageList.MessageEmail, messageList.MessageSubject, messageList.MessageContent, messageList.MessageId)
		if ret > 0 {
			functions.Logger().Info("Successfully Send Email to API " + strconv.Itoa(messageList.MessageId) + "")
		} else {
			functions.Logger().Info("Failed Send Email to API " + strconv.Itoa(messageList.MessageId) + "")
		}
	}
	rows.Close()
}

func sendMail(to string, subject string, message string, id int) int {
	emailfrom := ""
	emailfrom = GetFValueByFKeyValue("config", "config_name", "email_lembaga", "config_value")
	cPort := "587"
	cPort = GetFValueByFKeyValue("config", "config_name", "email_smtp_port", "config_value")
	cHost := "BERSAMA KITA MAJU"
	cHost = GetFValueByFKeyValue("config", "config_name", "email_smtp_host", "config_value")
	emailPass := ""
	emailPass = GetFValueByFKeyValue("config", "config_name", "email_pass_lembaga", "config_value")

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", emailfrom)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", message)
	port, _ := strconv.Atoi(cPort)
	dialer := gomail.NewDialer(
		cHost,
		port,
		emailfrom,
		emailPass,
	)

	err := dialer.DialAndSend(mailer)
	if err != nil {
		functions.Logger().Error(err.Error())
		return 0
	}

	stmt, err := conn.Prepare("UPDATE outbox_email SET STATUS=1 where chat_id=?")
	if err != nil {
		functions.Logger().Error(err.Error())
		return 0
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		functions.Logger().Error(err.Error())
		return 0
	} else {
		functions.Logger().Info("Successfully UpdateStatus")
		return 1
	}
}

func GetFValueByFKeyValue(Table string, FieldKey string, FieldKeyValue string, FieldTarget string) string {
	sqlStatement := "SELECT ifnull(" + FieldTarget + ",'') " + FieldTarget + " from " + Table + " where " + FieldKey + " = '" + FieldKeyValue + "'"
	rows, err := database.ConnectDB().Query(sqlStatement)
	if err != nil {
		functions.Logger().Error(err.Error())
		return ""
	}
	defer database.ConnectDB().Close()
	var Field string
	for rows.Next() {
		err = rows.Scan(&Field)
		if err != nil {
			return ""
		}
	}
	rows.Close()
	return Field

}

func SendMessageWA(Hp string, Pesan string, Id int) int {
	functions.Logger().Info("Starting Send Message to WABLAS")
	b := map[string]string{"phone": Hp, "message": Pesan}
	request := gorequest.New()
	resp, _, _ := request.Post("https://wablas.com/api/send-message").
		Set("Content-Type", "application/json").
		Set("Authorization", GetFValueByFKeyValue("config", "config_name", "token_wa", "config_value")).
		Send(b).
		End()
	fmt.Println(resp.Body)
	if resp != nil {
		if resp.StatusCode == http.StatusOK {
			stmt, err := conn.Prepare("UPDATE outbox SET STATUS=1 where chat_id=?")
			if err != nil {
				functions.Logger().Error(err.Error())
				return 0
			}
			defer stmt.Close()
			_, err = stmt.Exec(Id)
			if err != nil {
				functions.Logger().Error(err.Error())
				return 0
			} else {
				functions.Logger().Info("Successfully UpdateStatus")
			}

			functions.Logger().Info("Successfully Send Message to WABLAS")
			return 1
		} else {
			functions.Logger().Error("An Error Occured")
			return 0
		}
	} else {
		functions.Logger().Error("An Error Occured")
		return 0
	}

}
