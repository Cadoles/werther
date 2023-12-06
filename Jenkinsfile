@Library('cadoles') _

// Utilisation du pipeline "standard"
// Voir https://forge.cadoles.com/Cadoles/Jenkins/src/branch/master/doc/tutorials/standard-make-pipeline.md
standardMakePipeline([
    'dockerfileExtension': '''
    RUN apt-get update \
        && apt-get install -y zip jq

    RUN wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz \
        && rm -rf /usr/local/go \
        && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

    ENV PATH="${PATH}:/usr/local/go/bin"
    ''',
    'hooks': [
        'pre-release': {
            // Login into docker registry
            sh '''
            make .mktools
            echo "$MKT_GITEA_RELEASE_PASSWORD" | docker login --username "$MKT_GITEA_RELEASE_USERNAME" --password-stdin reg.cadoles.com
            '''
        }
    ],
    // Use credentials to push images to registry and pubish gitea release
    'credentials': [
        usernamePassword(credentialsId: 'kipp-credentials', usernameVariable: 'MKT_GITEA_RELEASE_USERNAME', passwordVariable: 'MKT_GITEA_RELEASE_PASSWORD')
    ]
])