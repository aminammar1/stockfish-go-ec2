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
                sh 'docker run --rm -v "$WORKSPACE:/work" -w /work alpine:3.23 sh -c "rm -rf go"'
                sh 'docker run --rm --user $(id -u):$(id -g) -e HOME=/tmp -e GOPATH=/tmp/go -v "$WORKSPACE:/work" -w /work golang:1.25.5 sh -c "go mod download && go install github.com/swaggo/swag/cmd/swag@latest && /tmp/go/bin/swag init -g cmd/server/main.go -o docs --parseDependency=false --parseInternal=true && go test -v ./... -short"'
            }
        }

        stage('Build') {
            steps {
                sh 'docker run --rm -v "$WORKSPACE:/work" -w /work alpine:3.23 sh -c "rm -rf go"'
                sh '''docker run --rm --user $(id -u):$(id -g) \\
                    -e HOME=/tmp -e GOPATH=/tmp/go \\
                    -v "$WORKSPACE:/work" -w /work \\
                    golang:1.25.5 \\
                    sh -c "go mod download && \\
                           go install github.com/swaggo/swag/cmd/swag@latest && \\
                           /tmp/go/bin/swag init -g cmd/server/main.go -o docs --parseDependency=false --parseInternal=true && \\
                           CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o bin/server ./cmd/server && \\
                           CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o bin/cli ./cmd/cli"'''
            }
        }

        stage('Docker Build') {
            steps {
                sh "docker build -t ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ."
                sh "docker tag ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION} ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:latest"
            }
        }

        stage('Performance Test') {
            steps {
                script {
                    try {
                        // Start the service container in the background
                        // We map port 8080 to the host so K6 can access it via localhost:8080
                        sh """docker run -d --name stockfish-test-${BUILD_NUMBER} \\
                            -p 8080:8080 \\
                            -e SSH_HOST=\${SSH_HOST} \\
                            -e SSH_USER=\${SSH_USER} \\
                            -e SSH_PRIVATE_KEY=\${SSH_PRIVATE_KEY} \\
                            ${GHCR_REGISTRY}/${GHCR_USER}/${IMAGE_NAME}:${VERSION}"""

                        // Wait for service to be ready (health check)
                        sh "sleep 5"

                        // Run K6 tests
                        // K6 runs in host network mode to access localhost:8080 easily
                        sh 'docker run --rm --network host -v "$WORKSPACE/tests/performance:/scripts" grafana/k6:latest run --vus 5 --duration 10s -e BASE_URL=http://localhost:8080 /scripts/load_test.js'
                    } catch (Exception e) {
                        currentBuild.result = 'FAILURE'
                        error("Performance tests failed: ${e.message}")
                    } finally {
                        // Cleanup
                        sh "docker stop stockfish-test-${BUILD_NUMBER} || true"
                        sh "docker rm stockfish-test-${BUILD_NUMBER} || true"
                    }
                }
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
