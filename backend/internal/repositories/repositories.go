package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/natneam/crypto-wallet-app/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	dbClient *mongo.Client
}

func NewRepository(dbClient *mongo.Client) *Repository {
	return &Repository{dbClient: dbClient}
}

func (r *Repository) SaveTransaction(ctx context.Context, newTransaction *models.TransactionResult) (models.TransactionResult, error) {
	collection := r.dbClient.Database("walletdb").Collection("transactions")
	insertCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(insertCtx, newTransaction)
	if err != nil {
		return *newTransaction, fmt.Errorf("failed to insert transaction into database: %v", err)
	}
	newTransaction.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return *newTransaction, nil
}

func (r *Repository) SaveWallet(ctx context.Context, newWallet *models.Wallet) (models.Wallet, error) {
	collection := r.dbClient.Database("walletdb").Collection("wallets")
	insertCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(insertCtx, newWallet)
	if err != nil {
		return *newWallet, fmt.Errorf("failed to insert wallet into database: %v", err)
	}
	newWallet.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return *newWallet, nil
}

func (r *Repository) GetWallet(ctx context.Context, address string, userId string) (models.Wallet, error) {
	collection := r.dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var wallet models.Wallet
	err := collection.FindOne(ctx, bson.M{"publickey": address, "userid": userId}).Decode(&wallet)
	if err != nil {
		return wallet, fmt.Errorf("failed to find wallet: %v", err)
	}
	return wallet, nil
}

func (r *Repository) ListWallets(ctx context.Context, userId string) ([]models.Wallet, error) {
	collection := r.dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"userid": userId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []models.Wallet
	if err := cursor.All(ctx, &wallets); err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *Repository) CreateUser(user *models.User) error {
	collection := r.dbClient.Database("walletdb").Collection("users")
	insertCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(insertCtx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				return fmt.Errorf("email already exists")
			} else if strings.Contains(err.Error(), "username") {
				return fmt.Errorf("username already exists")
			}
			return fmt.Errorf("duplicate key error")
		}
		return fmt.Errorf("failed to insert user into database: %v", err)
	}
	return nil
}

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	collection := r.dbClient.Database("walletdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	return &user, nil
}
