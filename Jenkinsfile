pipeline {
    agent any

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
        stage('Diagnostics') {
            steps {
                sh 'docker version'
                sh 'docker run --rm golang:1.25.5 go version'
            }
        }

        stage('Unit Test') {
            steps {
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "go test -v ./... -short" || true'
            }
        }

        stage('Performance Test') {
            steps {
                sh 'docker run --rm -v "$WORKSPACE/tests/performance:/scripts" grafana/k6:latest run --vus 5 --duration 10s /scripts/load_test.js || true'
            }
        }

        stage('Build') {
            steps {
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "go mod download"'
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "go install github.com/swaggo/swag/cmd/swag@latest"'
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "swag init -g cmd/server/main.go -o docs"'
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o bin/server ./cmd/server"'
                sh 'docker run --rm --user $(id -u):$(id -g) -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o bin/cli ./cmd/cli"'
            }
        }

        stage('Docker Build') {
            steps {
                sh "docker build -t ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ."
                sh "docker tag ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:latest"
            }
        }

        stage('Push to GHCR') {
            steps {
                withCredentials([string(credentialsId: 'ghcr-token', variable: 'GHCR_TOKEN')]) {
                    sh 'echo $GHCR_TOKEN | docker login ghcr.io -u aminammar1 --password-stdin'
                    sh "docker push ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION}"
                    sh "docker push ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:latest"
                }
            }
        }

        stage('Deploy to Render') {
            steps {
                withCredentials([
                    string(credentialsId: 'render-api-token', variable: 'RENDER_API_TOKEN'),
                    string(credentialsId: 'render-service-id', variable: 'RENDER_SERVICE_ID')
                ]) {
                    sh '''
                        docker run --rm curlimages/curl:8.6.0 -sS \
                          -X POST "$RENDER_API_URL/services/$RENDER_SERVICE_ID/deploys" \
                          -H "Authorization: Bearer $RENDER_API_TOKEN" \
                          -H "Content-Type: application/json" \
                          -d "{}"
                    '''
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
