pipeline {
    agent {
        docker {
            image 'golang:1.25.5-alpine3.23'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }

    environment {
        GHCR_REGISTRY = 'ghcr.io'
        GHCR_USER = 'aminammar1'
        IMAGE_NAME = 'stockfish-ec2-service'
        VERSION = "${env.BUILD_NUMBER}"
        RENDER_API_URL = 'https://api.render.com/v1'
        RENDER_SERVICE_ID = credentials('render-service-id')
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
                sh 'go test -v ./internal/adapters/stockfish_ssh/... -run TestHealth || true'
                sh 'go test -v ./... -short || true'
            }
        }

        stage('Performance Test') {
            agent {
                docker {
                    image 'grafana/k6:latest'
                    reuseNode true
                }
            }
            steps {
                sh 'k6 run --vus 5 --duration 10s tests/performance/load_test.js || true'
            }
        }

        stage('Build') {
            steps {
                sh 'go mod download'
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
                withCredentials([string(credentialsId: 'render-api-token', variable: 'RENDER_API_TOKEN')]) {
                    sh """
                        curl -X POST "${RENDER_API_URL}/services/${RENDER_SERVICE_ID}/deploys" \
                            -H "Authorization: Bearer ${RENDER_API_TOKEN}" \
                            -H "Content-Type: application/json"
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
