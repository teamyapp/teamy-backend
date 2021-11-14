package repo

import (
	"database/sql"
	"errors"
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

var sqlTaskStatues = map[entity.TaskStatus]int{
	entity.TaskStatusUpcoming:   1,
	entity.TaskStatusInProgress: 2,
	entity.TaskStatusDelivered:  3,
}

type Task interface {
	FindTasksForTeam(teamID oneEntity.ID, taskStatus entity.TaskStatus) ([]entity.Task, error)
	FindTasksForUser(userID oneEntity.ID, teamID oneEntity.ID, taskStatus entity.TaskStatus) ([]entity.Task, error)
	FindTaskNeedAttentionForUser(userID oneEntity.ID, teamID oneEntity.ID) (*entity.Task, error)
	FindTaskByID(taskID oneEntity.ID) (entity.Task, error)
	CreateTask(task entity.Task) error
}

type SQLTask struct {
	db *sql.DB
}

var _ Task = (*SQLTask)(nil)

func (S SQLTask) CreateTask(task entity.Task) error {
	statement := `
	INSERT INTO task(
		 goal,
		 due_at,
		 context,
		 owner_user_id,
		 work_scope_index,
		 effort,
		 num_of_unknowns
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
`
	_, err := S.db.Exec(statement, task.Goal, task.DueAt, task.Context, task.OwnerUserId, task.WorkScopeIndex, task.Effort, task.NumOfUnknowns)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (S SQLTask) FindTaskByID(taskID oneEntity.ID) (entity.Task, error) {
	ts := entity.Task{}
	err := S.db.QueryRow(`
	SELECT
	       id,
	       goal,
	       due_at,
	       context,
	       owner_user_id,
	       work_scope_index,
	       effort,
	       num_of_unknowns,
	       created_at,
	       updated_at
	FROM task
	WHERE id = $1;
`, int(taskID)).
		Scan(
			&ts.ID,
			&ts.Goal,
			&ts.DueAt,
			&ts.Context,
			&ts.OwnerUserId,
			&ts.WorkScopeIndex,
			&ts.Effort,
			&ts.NumOfUnknowns,
			&ts.CreatedAt,
			&ts.UpdatedAt)
	if err != nil {
		log.Println(err)
	}
	return ts, err
}

func (S SQLTask) FindTasksForTeam(teamID oneEntity.ID, taskStatus entity.TaskStatus) ([]entity.Task, error) {
	rows, err := S.db.Query(`
SELECT
       task.id,
       task.goal,
       task.due_at,
       task.context,
       task.owner_user_id,
       task.work_scope_index,
       task.effort,
       task.num_of_unknowns,
       task.created_at,
       task.updated_at
FROM team_task
INNER JOIN task ON team_task.task_id = task.id
WHERE team_id = $1
  AND task_status = $2;`,
		int(teamID), sqlTaskStatues[taskStatus])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	tasks := make([]entity.Task, 0)
	for rows.Next() {
		ts := entity.Task{}
		err = rows.Scan(
			&ts.ID,
			&ts.Goal,
			&ts.DueAt,
			&ts.Context,
			&ts.OwnerUserId,
			&ts.WorkScopeIndex,
			&ts.Effort,
			&ts.NumOfUnknowns,
			&ts.CreatedAt,
			&ts.UpdatedAt)
		if err != nil {
			log.Println(err)
			return tasks, err
		}

		tasks = append(tasks, ts)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return tasks, err
	}

	return tasks, nil
}

func (S SQLTask) FindTasksForUser(userID oneEntity.ID, teamID oneEntity.ID, taskStatus entity.TaskStatus) ([]entity.Task, error) {
	rows, err := S.db.Query(`
SELECT
       task.id,
       task.goal,
       task.due_at,
       task.context,
       task.owner_user_id,
       task.work_scope_index,
       task.effort,
       task.num_of_unknowns,
       task.created_at,
       task.updated_at
FROM team_task
INNER JOIN task ON team_task.task_id = task.id
WHERE team_task.team_id = $1
  AND team_task.task_status = $2
  AND task.owner_user_id = $3;`,
		int(teamID), sqlTaskStatues[taskStatus], int(userID))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	tasks := make([]entity.Task, 0)
	for rows.Next() {
		ts := entity.Task{}
		err = rows.Scan(
			&ts.ID,
			&ts.Goal,
			&ts.DueAt,
			&ts.Context,
			&ts.OwnerUserId,
			&ts.WorkScopeIndex,
			&ts.Effort,
			&ts.NumOfUnknowns,
			&ts.CreatedAt,
			&ts.UpdatedAt)
		if err != nil {
			log.Println(err)
			return tasks, err
		}

		tasks = append(tasks, ts)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return tasks, err
	}

	return tasks, nil
}

func (S SQLTask) FindTaskNeedAttentionForUser(userID oneEntity.ID, teamID oneEntity.ID) (*entity.Task, error) {
	row := S.db.QueryRow(`
SELECT
       task.id,
       task.goal,
       task.due_at,
       task.context,
       task.owner_user_id,
       task.work_scope_index,
       task.effort,
       task.num_of_unknowns,
       task.created_at,
       task.updated_at
FROM team_member
INNER JOIN task ON team_member.need_attention_task_id = task.id
WHERE team_member.user_id = $1
  AND team_member.team_id = $2;
`, int(userID), int(teamID))
	task := entity.Task{}
	err := row.Scan(
		&task.ID,
		&task.Goal,
		&task.DueAt,
		&task.Context,
		&task.OwnerUserId,
		&task.WorkScopeIndex,
		&task.Effort,
		&task.NumOfUnknowns,
		&task.CreatedAt,
		&task.UpdatedAt)
	if err == nil {
		return &task, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return nil, err
}

func NewSQLTask(db *sql.DB) SQLTask {
	return SQLTask{db: db}
}
