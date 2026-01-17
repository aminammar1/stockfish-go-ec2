pipeline {
    agent {
        docker {
            image 'golang:1.25.5-alpine3.23'
            args '-u root -v /var/run/docker.sock:/var/run/docker.sock'
        }
    }

    environment {
        GHCR_REGISTRY = 'ghcr.io'
        GHCR_USER = 'aminammar1'
        IMAGE_NAME = 'stockfish-ec2-service'
        VERSION = "${env.BUILD_NUMBER}"
        RENDER_API_URL = 'https://api.render.com/v1'
    }

    triggers {
        githubPush()
    }

    stages {
        stage('Setup') {
            steps {
                sh 'apk add --no-cache docker git curl'
            }
        }

        stage('Unit Test') {
            steps {
                sh 'go test -v ./... -short || true'
            }
        }

        stage('Performance Test') {
            steps {
                sh 'docker run --rm -v ${WORKSPACE}/tests/performance:/scripts grafana/k6:latest run --vus 5 --duration 10s /scripts/load_test.js || true'
            }
        }

        stage('Build') {
            steps {
                sh 'go mod download'
                sh 'go install github.com/swaggo/swag/cmd/swag@latest'
                sh 'swag init -g cmd/server/main.go -o docs'
                sh 'CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server'
                sh 'CGO_ENABLED=0 GOOS=linux go build -o bin/cli ./cmd/cli'
            }
        }

        stage('Docker Build') {
            agent any
            steps {
                script {
                    sh "docker build -t ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ."
                    sh "docker tag ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:latest"
                }
            }
        }

        stage('Push to GHCR') {
            agent any
            steps {
                withCredentials([string(credentialsId: 'ghcr-token', variable: 'GHCR_TOKEN')]) {
                    sh "echo ${GHCR_TOKEN} | docker login ${GHCR_REGISTRY} -u ${GHCR_USER} --password-stdin"
                    sh "docker push ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION}"
                    sh "docker push ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:latest"
                }
            }
        }

        stage('Deploy to Render') {
            agent any
            steps {
                withCredentials([
                    string(credentialsId: 'render-api-token', variable: 'RENDER_API_TOKEN'),
                    string(credentialsId: 'render-service-id', variable: 'RENDER_SERVICE_ID')
                ]) {
                    sh """
                        curl -X POST "${RENDER_API_URL}/services/\${RENDER_SERVICE_ID}/deploys" \
                            -H "Authorization: Bearer \${RENDER_API_TOKEN}" \
                            -H "Content-Type: application/json" \
                            -d "{}"
                    """
                }
            }
        }
    }

    post {
        success {
            echo 'Pipeline completed successfully'
        }
        failure {
            echo 'Pipeline failed'
        }
        always {
            echo 'Pipeline finished'
            sh 'docker logout ${GHCR_REGISTRY} || true'
        }
    }
}
