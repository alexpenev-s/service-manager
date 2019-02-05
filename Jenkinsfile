pipeline {
    agent {
        kubernetes {
            label "${UUID.randomUUID().toString()}"
            defaultContainer 'jdk'
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: jdk
    image: openjdk:8
    command:
    - cat
    tty: true
  - name: jnlp
    image: docker.wdf.sap.corp:50001/sap-production/jnlp-alpine:3.26.1-sap-02
    args: ['\$(JENKINS_SECRET)', '\$(JENKINS_NAME)']
"""
        }
    }
    stages {
        stage('Build & Check') {
            steps {
                sh 'java -version'
            }
        }
    }
}