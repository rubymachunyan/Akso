pipeline {
    agent {
        label 'slc03rpq_jenkins_slaves_akso'
    }
	
    stages {	
	stage('build') {
            steps {
                echo 'Begin to build Akso......'
		sh 'make prepare'
		sh 'make build'
                echo 'Build Akso Done......'
            }
        }
    }
}
