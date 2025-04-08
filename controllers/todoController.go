package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jeffthorne/tasky/auth"
	"github.com/jeffthorne/tasky/database"
	"github.com/jeffthorne/tasky/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var todoCollection *mongo.Collection = database.OpenCollection(database.Client, "todos")

func GetTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	id := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
		return
	}

	var todo models.Todo
	err = todoCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func ClearAll(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userid := c.Param("userid")
	if userid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	result, err := todoCollection.DeleteMany(ctx, bson.M{"userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      "All todos deleted",
		"deletedCount": result.DeletedCount,
	})
}

func GetTodos(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userid := c.Param("userid")
	findResult, err := todoCollection.Find(ctx, bson.M{"userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}
	defer findResult.Close(ctx)

	var todos []models.Todo
	for findResult.Next(ctx) {
		var todo models.Todo
		if err := findResult.Decode(&todo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode todo"})
			return
		}
		todos = append(todos, todo)
	}

	if err := findResult.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while fetching todos"})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func DeleteTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	id := c.Param("id")
	userid := c.Param("userid")

	if id == "" || userid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todo ID and User ID are required"})
		return
	}

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
		return
	}

	deleteResult, err := todoCollection.DeleteOne(ctx, bson.M{"_id": objId, "userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found or not owned by user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      "Todo deleted successfully",
		"deletedCount": deleteResult.DeletedCount,
	})
}

func UpdateTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var newTodo models.Todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if newTodo.ID.IsZero() || newTodo.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todo ID and User ID are required"})
		return
	}

	result, err := todoCollection.UpdateOne(
		ctx,
		bson.M{"_id": newTodo.ID, "userid": newTodo.UserID},
		bson.M{"$set": newTodo},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found or not owned by user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       "Todo updated successfully",
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
	})
}

func AddTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userid := c.Param("userid")
	if userid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var todo models.Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	todo.ID = primitive.NewObjectID()
	todo.UserID = userid

	result, err := todoCollection.InsertOne(ctx, todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": "Todo created successfully",
		"id":      result.InsertedID,
	})
}
