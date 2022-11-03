@Library('cadoles') _

pipeline {
  agent {
    dockerfile {
      label 'docker'
      filename 'Dockerfile'
      dir 'misc/ci'
    }
  }

  stages {
    stage('Build and publish packages') {
      when {
        anyOf {
          branch 'master'
          branch 'develop'
        }
      }
      steps {
        script {
          List<String> packagers = ['deb', 'rpm']
          packagers.each { pkgr ->
            sh "make NFPM_PACKAGER='${pkgr}' build package"
          }

          List<String> attachments = sh(returnStdout: true, script: "find dist -type f -name '*.deb' -or -name '*.rpm' -or -name '*.ipk'").split(' ')
          String releaseVersion = sh(returnStdout: true, script: "git describe --always | rev | cut -d '/' -f 1 | rev").trim()

          String releaseBody = """
          _Publication automatisée réalisée par Jenkins._ [Voir le job](${env.RUN_DISPLAY_URL})
          """

          gitea.release('forge-jenkins', 'Cadoles', 'hydra-werther', [
            'attachments': attachments,
            'body': releaseBody,
            'releaseName': "${releaseVersion}",
            'releaseVersion': "${releaseVersion}"
          ])
        }
      }
    }
  }

  post {
    always {
      cleanWs()
    }
  }
}