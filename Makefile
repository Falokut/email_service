project_name = email_service

.docker-build:
	docker compose -f $(project_name).yml up --build