package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Response struct {
	Success int         `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Token   string      `json:"token,omitempty"`
	UserID  int         `json:"user_id,omitempty"`
}

/*|| ========================= REGISTER =================================== ||*/

// func isPhoneNumberValid(phoneNumber string) bool {
// 	pattern := `^08[0-9]{9,11}$`
// 	matched, _ := regexp.MatchString(pattern, phoneNumber)
// 	return matched
// }

func RegisterUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: err.Error(), Data: nil})
	}
	// TODO: new Validation for Notification Token required from FRONT END!!
	if user.FullName == "" || user.Password == "" || user.PhoneNumber == "" || user.Email == "" || user.NotificationToken == "" {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Fullname, Password, NoHP, and Email are required fields", Data: nil})
	}

	if isEmailExists(user.Email) {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Email is already registered", Data: nil})
	}

	// if !isPhoneNumberValid(user.PhoneNumber) {
	// 	return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Invalid phone number format", Data: nil})
	// }

	if isPhoneNumberExists(user.PhoneNumber) {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Phone number is already registered", Data: nil})
	}
	username := generateUsername(user.FullName)

	user.Role = "masyarakat"
	user.Username = username
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if user.PhotoProfile == "" {
		user.PhotoProfile = ""
	}
	if user.TempatLahir == "" {
		user.TempatLahir = ""
	}
	if user.TanggalLahir.IsZero() {
		user.TanggalLahir = time.Time{}
	}
	if user.JenisKelamin == "" {
		user.JenisKelamin = ""
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to hash password", Data: nil})
	}
	user.Password = hashedPassword

	userID, err := saveUserToDatabase(&user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to register user", Data: nil})
	}

	user.ID = uint(userID)

	return c.Status(http.StatusOK).JSON(Response{
		Success: 200,
		Message: "User registered successfully",
		Data:    user})
}

func isEmailExists(email string) bool {
	db := database.GetDBInstance()
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email)
	row.Scan(&count)
	return count > 0
}

func isPhoneNumberExists(phoneNumber string) bool {
	db := database.GetDBInstance()
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE phone_number = ?", phoneNumber)
	row.Scan(&count)
	return count > 0
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func saveUserToDatabase(user *models.User) (int64, error) {
	db := database.GetDBInstance()
	// ! Updated: added new query for adding Notification Token
	query := "INSERT INTO users (role, full_name, username, photo_profile, phone_number, email, password, notification_token, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query,
		user.Role,
		user.FullName,
		user.Username,
		user.PhotoProfile,
		user.PhoneNumber,
		user.Email,
		user.Password,
		user.NotificationToken,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		log.Println("Error saving user to database:", err)
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting last inserted ID:", err)
		return 0, err
	}

	log.Printf("User saved successfully to database with ID: %d\n", userID)
	return userID, nil
}

func generateUsername(fullName string) string {
	var username string
	db := database.GetDBInstance()

	names := strings.Fields(fullName)
	firstName := names[0]

	for {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(90000000) + 10000000
		username = fmt.Sprintf("%s%d", strings.Title(strings.ToLower(firstName)), randomNumber)
		if !isUsernameExists(db, username) {
			break
		}
	}
	return username
}

func isUsernameExists(db *sql.DB, username string) bool {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username)
	row.Scan(&count)
	return count > 0
}

/*||============================== LOGIN =================================== ||*/
func LoginUser(c *fiber.Ctx) error {
	var credentials models.LoginCredentials
	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: err.Error(), Data: nil, UserID: 0})
	}

	user, err := getUserByCredentials(credentials)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(Response{Success: 0, Message: "Email atau Username, oataur Phone Number or password salah", Data: nil, UserID: 0})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(Response{Success: 0, Message: "Email atau Username, oataur Phone Number or password salah", Data: nil, UserID: 0})
	}

	// Tidak ada pemanggilan fungsi VerifyToken di sini

	token, err := generateAuthToken(int64(user.ID), user.Role)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to generate token", Data: nil, UserID: 0})
	}

	fullUser, err := getUserByID(int(user.ID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to fetch user details", Data: nil, UserID: 0})
	}

	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Anda Berhasil Login", Data: fullUser, Token: token})
}

func getUserByCredentials(credentials models.LoginCredentials) (models.User, error) {
	db := database.GetDBInstance()

	var user models.User
	query := "SELECT id, username, email, phone_number, role, password FROM users WHERE email = ? OR username = ? OR phone_number = ?"
	err := db.QueryRow(query, credentials.Email, credentials.Username, credentials.PhoneNumber).Scan(&user.ID, &user.Username, &user.Email, &user.PhoneNumber, &user.Role, &user.Password)

	if err != nil {
		log.Println("Error getting user by credentials:", err)
		return models.User{}, err
	}
	return user, nil
}

func generateAuthToken(userID int64, role string) (string, error) {
	error := godotenv.Load()
	if error != nil {
		panic("Cannot Find ENV file")
	}
	expirationTime := time.Now().Add(10 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
		"role":    role,
	}

	jwt_secreet := os.Getenv("JWT_SECRET_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwt_secreet))

	if err != nil {
		log.Println("Error generating JWT token:", err)
		return "", err
	}

	return signedToken, nil
}

func getUserByID(userID int) (models.User, error) {
	db := database.GetDBInstance()
	var photoProfile sql.NullString
	// !UUPDATED: field query for retireve Notification Token
	var user models.User
	query := "SELECT id, full_name, username, role, photo_profile, phone_number, email, password, notification_token, created_at, updated_at FROM users WHERE id = ?"
	err := db.QueryRowContext(context.Background(), query, userID).Scan(
		&user.ID,
		&user.FullName,
		&user.Username,
		&user.Role,
		&photoProfile,
		&user.PhoneNumber,
		&user.Email,
		&user.Password,
		&user.NotificationToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if photoProfile.Valid {
		user.PhotoProfile = photoProfile.String
	} else {
		user.PhotoProfile = ""
	}
	if err != nil {
		log.Println("Error getting user by ID:", err)
		return models.User{}, err
	}
	fmt.Println("Recived User Dat:", user)
	return user, nil
}
