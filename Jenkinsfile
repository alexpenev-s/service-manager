pipeline {
    def label = "pod-${UUID.randomUUID().toString()}"
    podTemplate(label: label) {
        node(label) {
            stage('Run shell') {
                sh 'echo hello world'
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
