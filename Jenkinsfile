pipeline {
    agent {
        def label = "mypod-${UUID.randomUUID().toString()}"
        podTemplate(label: label) {
            node(label) {
                stage('Run shell') {
                    sh 'echo hello world'
                }
            }
        }
    }
    stages {
        stage('build') {
            steps {
                sh 'npm --version'
            }
        }
    }
}
