package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"middleproject/scripts"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Advpost struct {
	PostID      int      `json:"id"`
	Title       string   `json:"title"`
	Uname       string   `json:"uname"`
	Uid         string   `json:"uid"`
	Uimge       string   `json:"uimage"`
	Time        string   `json:"time"`
	Subject     []string `json:"subjects"`
	SomeContent string   `json:"somecontent"`
	Islike      bool     `json:"islike"`
	Iscollect   bool     `json:"iscollect"`
	Likenum     int      `json:"likenum"`
	Collectnum  int      `json:"collectnum"`
}

// 推荐逻辑设计
func AdvisePost(uid int, page int, isattention string) ([]Advpost, error, int) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		return nil, err_conn, 0
	}
	defer db.Close()
	var posts []Advpost
	if isattention == "true" {
		//获取关注的人的帖子，按喜欢数量排序
		query := "SELECT posts.post_id,posts.title,users.Uname,users.user_id,users.avatar,posts.publish_time,posts.post_subject,posts.content,posts.like_count,posts.favorite_count from posts,users,userfollows where posts.user_id=users.user_id AND posts.user_id=userfollows.followed_id AND userfollows.follower_id=? order by posts.view_count"
		rows, err_query := db.Query(query, uid)
		if err_query != nil {
			fmt.Println(err_query.Error())
			return posts, err_query, 0
		}
		for rows.Next() {
			var post Advpost

			var subject sql.NullString
			var uidint int
			err_scan := rows.Scan(&post.PostID, &post.Title, &post.Uname, &uidint, &post.Uimge, &post.Time, &subject, &post.SomeContent, &post.Likenum, &post.Collectnum)
			if err_scan != nil {
				fmt.Println(err_scan.Error())
				return posts, err_scan, 0
			}
			post.Time = post.Time[0 : len(post.Time)-3]
			post.Uid = strconv.Itoa(uidint)
			var err_url error
			err_url, post.Uimge = scripts.GetUrl(post.Uimge)
			if err_url != nil {
				return posts, err_url, 0
			}
			if subject.Valid {
				str := subject.String
				post.Subject = strings.Split(str[1:len(str)-1], ",")
				//去除双引号
				for i := 0; i < len(post.Subject); i++ {
					if i == 0 {
						post.Subject[i] = post.Subject[i][1 : len(post.Subject[i])-1]
					} else {
						post.Subject[i] = post.Subject[i][2 : len(post.Subject[i])-1]
					}
				}

			} else {

				post.Subject = []string{"无关键字"}
			}
			//判断是否喜欢
			query = "select liker_id from postlikes where liker_id=? and post_id=?"
			row := db.QueryRow(query, uid, post.PostID)
			var like_id int
			err_scan = row.Scan(&like_id)
			if err_scan != nil {
				post.Islike = false
			} else {
				post.Islike = true
			}
			//判断是否收藏
			query = "select user_id from PostFavorites where user_id=? and post_id=?"
			row = db.QueryRow(query, uid, post.PostID)
			var favorite_id int
			err_scan = row.Scan(&favorite_id)
			if err_scan != nil {
				post.Iscollect = false
			} else {
				post.Iscollect = true
			}
			if len(post.SomeContent) > 300 {
				post.SomeContent = post.SomeContent[0:300]
				post.SomeContent = post.SomeContent + "..."
			}
			posts = append(posts, post)

		}

	} else {
		//获取所有的帖子，按喜欢数量排序
		query := "SELECT posts.post_id,posts.title,users.Uname,users.user_id,users.avatar,posts.publish_time,posts.post_subject,posts.content,posts.like_count,posts.favorite_count from posts,users where posts.user_id=users.user_id order by posts.view_count"
		rows, err_query := db.Query(query)
		if err_query != nil {
			fmt.Println(err_query.Error())
			return posts, err_query, 0
		}
		for rows.Next() {
			var post Advpost
			var subject sql.NullString
			var uidint int
			err_scan := rows.Scan(&post.PostID, &post.Title, &post.Uname, &uidint, &post.Uimge, &post.Time, &subject, &post.SomeContent, &post.Likenum, &post.Collectnum)
			post.Time = post.Time[0 : len(post.Time)-3]
			if err_scan != nil {
				fmt.Println(err_scan.Error())
				return posts, err_scan, 0
			}
			post.Uid = strconv.Itoa(uidint)
			var err_url error
			err_url, post.Uimge = scripts.GetUrl(post.Uimge)
			if err_url != nil {
				return posts, err_url, 0
			}
			if subject.Valid {
				str := subject.String
				post.Subject = strings.Split(str[1:len(str)-1], ",")
				//去除双引号
				for i := 0; i < len(post.Subject); i++ {
					if i == 0 {
						post.Subject[i] = post.Subject[i][1 : len(post.Subject[i])-1]
					} else {
						post.Subject[i] = post.Subject[i][2 : len(post.Subject[i])-1]
					}
				}
			} else {
				post.Subject = []string{"无关键字"}
			}
			fmt.Println("uid:", uid)

			//判断是否喜欢
			if uid == -1 {
				post.Islike = false
				post.Iscollect = false
			} else {
				query = "select liker_id from postlikes where liker_id=? and post_id=?"
				row := db.QueryRow(query, uid, post.PostID)
				var like_id int
				err_scan = row.Scan(&like_id)
				if err_scan != nil {
					post.Islike = false
				} else {
					post.Islike = true
				}
				//判断是否收藏
				query = "select user_id from PostFavorites where user_id=? and post_id=?"
				row = db.QueryRow(query, uid, post.PostID)
				var favorite_id int
				err_scan = row.Scan(&favorite_id)
				if err_scan != nil {
					post.Iscollect = false
				} else {
					post.Iscollect = true
				}
			}
			if len(post.SomeContent) > 300 {
				post.SomeContent = post.SomeContent[0:300]
				post.SomeContent = post.SomeContent + "..."
			}
			fmt.Println(post)
			posts = append(posts, post)
		}
	}
	var realPost []Advpost
	for i := 0; i < len(posts); i++ {
		if i >= (page-1)*10 && i < page*10 {
			realPost = append(realPost, posts[i])
		}
	}
	totalpage := len(posts) / 10
	if len(posts)%10 != 0 {
		totalpage++
	}
	return realPost, nil, totalpage
}

// 发帖子接口
func PublishPost(c *gin.Context) {
	var data model.Post
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "发帖绑定请求数据失败"})
		return
	}
	uidstr := c.DefaultQuery("uid", "-1")

	uid, err := strconv.Atoi(uidstr)
	if err != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}
	data.UserID = uid
	erro, msg, idstr := data.AddPost()
	if erro != nil {
		fmt.Println(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "logid": idstr})

}

// 获得推荐帖子
func GetRecommendPost(c *gin.Context) {
	var pagestr string = c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pagestr)

	var isattention string = c.DefaultQuery("isattion", "false")
	var uidstr = c.DefaultQuery("uid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	var posts []Advpost
	if err_uid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": -1})
		return
	}
	posts, err_adv, num := AdvisePost(uid, page, isattention)

	fmt.Println(posts)
	if err_adv != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": num})
	return

}

// 帖子内部评论结构体
type PostComment struct {
	CID      string `json:"id"`
	UID      string `json:"uid"`
	Content  string `json:"content"`
	UName    string `json:"uname"`
	UImage   string `json:"uimage"`
	Time     string `json:"time"`
	IsLike   bool   `json:"islike"`
	Likenum  int    `json:"likenum"`
	Replynum int    `json:"replynum"`
	Replies  string `json:"replies"`
}

// PostData 定义帖子结构体
type PostData struct {
	Title      string        `json:"title"`
	Subjects   []string      `json:"subjects"`
	Content    string        `json:"content"`
	ImageNum   int           `json:"imagenum"`
	UID        string        `json:"uid"`
	UName      string        `json:"uname"`
	UImage     string        `json:"uimage"`
	IsAttion   bool          `json:"isattion"`
	Time       string        `json:"time"`
	IsLike     bool          `json:"islike"`
	IsCollect  bool          `json:"iscollect"`
	ViewNum    int           `json:"viewnum"`
	LikeNum    int           `json:"likenum"`
	CollectNum int           `json:"collectnum"`
	ComNum     int           `json:"comnum"`
	Comments   []PostComment `json:"comments"`
}

// 获取评论信息
func GetCommentInfo(page_num int, postid int, uid int, comid int) (error, []PostComment) {
	var comments []PostComment
	db, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, comments
	}
	defer db.Close()
	//按照喜欢数量排序，10条
	var query string
	var rows *sql.Rows
	if comid == -1 {
		query = "select comments.comment_id, users.Uname,comments.content,users.user_id,users.avatar,comments.comment_time,comments.like_count,comments.reply_count from users,comments where users.user_id=comments.commenter_id AND comments.post_id=? AND parent_comment_id is null order by comments.like_count desc limit ?,10"
		var err_query error
		rows, err_query = db.Query(query, postid, page_num)
		if err_query != nil {
			return err_query, comments
		}
	} else {
		query = "select comments.comment_id, users.Uname,comments.content,users.user_id,users.avatar,comments.comment_time,comments.like_count,comments.reply_count from users,comments where users.user_id=comments.commenter_id AND comments.post_id=? AND parent_comment_id = ? order by comments.like_count desc limit ?,5"
		var err_query2 error
		rows, err_query2 = db.Query(query, postid, comid, page_num)
		if err_query2 != nil {
			return err_query2, comments
		}
	}

	for rows.Next() {
		var comment PostComment
		var uid int
		var cid int
		err_scan := rows.Scan(&cid, &comment.UName, &comment.Content, &uid, &comment.UImage, &comment.Time, &comment.Likenum, &comment.Replynum)
		if err_scan != nil {
			return err_scan, comments
		}
		comment.UID = strconv.Itoa(uid)
		comment.CID = strconv.Itoa(cid)
		var err_url error
		err_url, comment.UImage = scripts.GetUrl(comment.UImage)
		if err_url != nil {
			return err_url, comments
		}
		if uid == -1 {
			comment.IsLike = false
		} else {
			query = "select liker_id from commentlikes where liker_id=? and comment_id=?"
			row := db.QueryRow(query, uid, cid)
			var like_id int
			err_scan = row.Scan(&like_id)
			if err_scan != nil {
				comment.IsLike = false
			} else {
				comment.IsLike = true
			}
		}
		query = "select content from comments where parent_comment_id=? order by like_count limit 1"
		row := db.QueryRow(query, cid)
		var reply string
		err_scan = row.Scan(&reply)
		if err_scan != nil {
			comment.Replies = ""
		} else {
			comment.Replies = reply
		}
		comments = append(comments, comment)
	}

	return nil, comments

}

// 获取帖子信息
func GetPostInfo(c *gin.Context) {
	postidstr := c.DefaultQuery("id", "-1")
	postid, err_tran := strconv.Atoi(postidstr)
	if err_tran != nil || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	Uidstr := c.DefaultQuery("uid", "-1")
	Uid_P, err_tran := strconv.Atoi(Uidstr)

	if err_tran != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	var post PostData

	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	defer db.Close()
	//查询帖子信息
	query := "select title,post_subject,content,images,user_id,publish_time,view_count,like_count,favorite_count,comment_count from posts where post_id=?"
	row := db.QueryRow(query, postid)
	var subject sql.NullString
	var images sql.NullString
	var uid int
	err_scan := row.Scan(&post.Title, &subject, &post.Content, &images, &uid, &post.Time, &post.ViewNum, &post.LikeNum, &post.CollectNum, &post.ComNum)
	if err_scan != nil {
		fmt.Println(err_scan.Error())
		c.JSON(404, gin.H{})
		return
	}
	post.UID = strconv.Itoa(uid)
	if subject.Valid {
		str := subject.String
		post.Subjects = strings.Split(str[1:len(str)-1], ",")
		//去除双引号
		for i := 0; i < len(post.Subjects); i++ {
			if i == 0 {
				post.Subjects[i] = post.Subjects[i][1 : len(post.Subjects[i])-1]
			} else {
				post.Subjects[i] = post.Subjects[i][2 : len(post.Subjects[i])-1]
			}
		}
	} else {
		post.Subjects = []string{}
	}
	if images.Valid {
		str := images.String
		if str == "[]" {
			post.ImageNum = 0
		} else {
			post.ImageNum = strings.Count(str, ",") + 1
		}
	} else {
		post.ImageNum = 0
	}
	//查询用户信息
	query = "select Uname,avatar from users where user_id=?"
	row = db.QueryRow(query, uid)
	err_scan = row.Scan(&post.UName, &post.UImage)
	if err_scan != nil {
		fmt.Println(err_scan.Error())
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var err_url error
	err_url, post.UImage = scripts.GetUrl(post.UImage)
	if err_url != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if Uid_P != -1 {
		query = "select follower_id from userfollows where follower_id=? and followed_id=?"
		row = db.QueryRow(query, Uid_P, uid)
		var follower_id int
		err_scan = row.Scan(&follower_id)
		if err_scan != nil {
			post.IsAttion = false
		} else {
			post.IsAttion = true
		}
		query = "select liker_id from postlikes where liker_id=? and post_id=?"
		row = db.QueryRow(query, Uid_P, postid)
		var like_id int
		err_scan = row.Scan(&like_id)
		if err_scan != nil {
			post.IsLike = false
		} else {
			post.IsLike = true
		}
		query = "select user_id from PostFavorites where user_id=? and post_id=?"
		row = db.QueryRow(query, Uid_P, postid)
		var favorite_id int
		err_scan = row.Scan(&favorite_id)
		if err_scan != nil {
			post.IsCollect = false
		} else {
			post.IsCollect = true
		}
	} else {
		post.IsAttion = false
		post.IsLike = false
		post.IsCollect = false
	}
	var err_get error
	err_get, post.Comments = GetCommentInfo(0, postid, Uid_P, -1)
	if err_get != nil {
		fmt.Println(err_get.Error())
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	query = "update posts set view_count=view_count+1 where post_id=?"
	_, err_update := db.Exec(query, postid)
	if err_update != nil {
		fmt.Println(err_update.Error())
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, post)

}

func GetPostImage(c *gin.Context) {
	postIDstr := c.DefaultQuery("logid", "-1")
	imagenumstr := c.DefaultQuery("imageid", "-1")
	postID, err_tran := strconv.Atoi(postIDstr)
	if err_tran != nil || postID == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	imagenum, err_tran := strconv.Atoi(imagenumstr)
	if err_tran != nil || imagenum == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	defer db.Close()
	query := "select images from posts where post_id=?"
	row := db.QueryRow(query, postID)
	var images sql.NullString
	err_scan := row.Scan(&images)
	if err_scan != nil {
		fmt.Println(err_scan.Error())
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if images.Valid {
		str := images.String
		if str == "[]" {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		str = str[1 : len(str)-1]
		image := strings.Split(str, ",")
		fmt.Println(image)
		fmt.Println(image[0])

		if imagenum >= len(image) {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		if imagenum == 0 {
			image[imagenum] = image[imagenum][1 : len(image[imagenum])-1]
		} else {
			image[imagenum] = image[imagenum][2 : len(image[imagenum])-1]
		}
		err_url, url := scripts.GetUrl(image[imagenum])
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

}
