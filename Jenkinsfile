pipeline {

    agent any

    environment {
        IMAGE_NAME = 'quay.io/csye7125_webapp/webapp-container-registry/kafka-producer'
    }

    triggers {
        GenericTrigger(
             genericVariables: [
                 [key: 'LATEST_RELEASE_TAG', value: '$.release.tag_name']
             ],
             token: 'kafka-producer-token'
        )
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm: [
                    $class: 'GitSCM',
                    branches: [[name: '*/master']],
                    doGenerateSubmoduleConfigurations: false,
                    extensions: [[$class: 'CloneOption', noTags: false]],
                    submoduleCfg: [],
                    userRemoteConfigs: [[credentialsId: 'GITHUB_TOKEN',
                    url: 'https://github.com/cyse7125-fall2023-group01/kafka-producer.git']]
                ]
            }
        }
        stage('Print Tag Details') {
            steps {
                script {
                    env.LATEST_RELEASE_TAG = sh(returnStdout: true, script: 'git describe --tags --abbrev=0').trim()
                    echo "Latest Release Tag: ${env.LATEST_RELEASE_TAG}"
                }
            }
        }
        stage('Docker Version') {
            steps {
                script {
                    sh 'docker version'
                }
            }
        }
        stage('Build Docker Image') {
            steps {
                script {
                    sh "docker build --no-cache -t ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG} ."
                    sh "docker image tag ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG} ${env.IMAGE_NAME}:latest"
                }
            }
        }
        stage('List Docker Images') {
            steps {
                script {
                    sh 'docker image ls'
                }
            }
        }
        stage('Quay Login') {
            steps {
                script {
                    withCredentials([
                    string(credentialsId: 'quayLogin', variable: 'quayLogin'),
                    string(credentialsId: 'quayEncryptedPwd', variable: 'quayEncryptedPwd')]) {
                        sh 'docker login -u=${quayLogin} -p=${quayEncryptedPwd} quay.io'
                    }
                }
            }
        }
        stage('Push Image To Quay') {
            steps {
                script {
                   withCredentials([string(credentialsId: 'quayEncryptedPwd', variable: 'quayEncryptedPwd')]) {
                        sh "docker push ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG}"
                        sh "docker push ${env.IMAGE_NAME}:latest"
                        sh "docker image rmi ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG}"
                    }
                }
            }
        }
    }
}
