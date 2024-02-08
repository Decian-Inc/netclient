/* groovylint-disable LineLength */
def majorVersion = ''
def minorVersion = ''
def patchVersion = ''
def dx_patchVersion = ''
def buildSkipped = false
def preIncVersion = ''
def postIncVersion = ''

def patchIncrementBranchPatterns = ['^dcx/main','^dcx/releases/.*']


pipeline {
    agent {
        label 'ubuntu22-vm'
    }

    tools { go 'Default'}

    options {
        disableConcurrentBuilds(abortPrevious: false)
    }
    environment {
        DOCKER_REGISTRY = 'nexus-registry.decian.net'
        IMAGE_NAME = 'dx-netclient'
        GIT_URL = 'git@github.com:Decian-Inc/netclient.git'
    }

    stages {
        stage('Skip?') {
        agent any
        steps {
            script {
                if (sh(script: "git log -1 --pretty=%B | fgrep -ie '[skip ci]' -e '[ci skip]'", returnStatus: true) == 0) {
                    def isManualTrigger = currentBuild.rawBuild.getCauses()[0].toString().contains('UserIdCause')
                    if (!isManualTrigger) {
                        currentBuild.result = 'SUCCESS'
                        currentBuild.description = 'Build skipped due to commit message'
                        buildSkipped = true
                        return
                    }
                }
            }
        }
        }
        stage('Checkout') {
            when {
                expression { return !buildSkipped }
            }
            steps {
                // checkout scm
                checkout changelog: false,
                    scm: scmGit(
                        branches: [[name: env.BRANCH_NAME]],
                        userRemoteConfigs: [[
                            credentialsId: 'jenkins-github-ssh-key',
                            url: env.GIT_URL ]]
                        )
            }
        }

        stage('Version Management') {
            steps {
                script {
                    preIncVersion = readFile("${env.WORKSPACE}/VERSION").trim()
                    (majorVersion, minorVersion, patchVersion, dx_patchVersion) = preIncVersion.tokenize('.')

                   // display version info
                    echo "Current Version: ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"
                    if (!buildSkipped) {
                        if (anyOf { patchIncrementBranchPatterns.collect { pattern ->
                            expression { env.BRANCH_NAME ==~ pattern }
                        } }) {
                            // Bump Patch Version, commit
                            dx_patchVersion = dx_patchVersion.toInteger() + 1
                            echo "New Version: ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"
                            sh "echo ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion} > VERSION"
                            postIncVersion = readFile("${env.WORKSPACE}/VERSION").trim()
                        }
                    }

                    currentBuild.displayName = "# ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion} | ${BRANCH_NAME}"

                }
            }
        }

        stage('Build Push Docker image') {
            when {
                expression { return !buildSkipped }
            }
            steps {
                script {
                    def baseNetclientVersion = "${majorVersion}.${minorVersion}.${patchVersion}"
                    def version = "${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"

                    // specify target platforms
                    def buildXPlatforms = []
                    buildXPlatforms.add("linux/amd64")
                    // buildXPlatforms.add("linux/arm64")
                    // buildXPlatforms.add("linux/arm/v7")

                    // specify docker tags
                    def dockerTags = []
                    dockerTags.add("${version}-${env.BRANCH_NAME.replaceAll("/", "-")}-${env.BUILD_NUMBER}")
                    dockerTags.add("${version}-${env.BRANCH_NAME.replaceAll("/", "-")}")

                    if (anyOf { patchIncrementBranchPatterns.collect { pattern ->
                            expression { env.BRANCH_NAME ==~ pattern }
                        } }) {
                        dockerTags.add("${version}")
                        dockerTags.add("${majorVersion}.${minorVersion}.${patchVersion}")
                        dockerTags.add("${majorVersion}.${minorVersion}")
                        dockerTags.add("${majorVersion}")
                    }

                    // build description
                    def descPlatformLIs= buildXPlatforms.collect { platform -> "<li>$platform</li>" }.join('')
                    def descTagLIs = dockerTags.collect { tag -> "<li>$DOCKER_REGISTRY/$IMAGE_NAME:${tag}</li>" }.join('')
                    currentBuild.description = """
                    <h1>Platforms</h1>
                    <ul>
                    ${descPlatformLIs}
                    </ul>
                    <br />
                    <h1>Docker Images</h1>
                    <ul>
                    ${descTagLIs}
                    </ul>
                    """



                    docker.withRegistry('https://nexus-registry.decian.net', 'nexus-docker-writer-username-password') {
                        def buildxCmdPlatforms = buildXPlatforms.join(',')
                        def buildxCmdTags = dockerTags.collect { tag -> "-t $DOCKER_REGISTRY/$IMAGE_NAME:${tag}" }.join(' ')
                        def buildxCmd = "docker buildx build --build-arg version=$baseNetclientVersion --platform ${buildxCmdPlatforms} --push $buildxCmdTags ."

                        sh """
                            docker buildx create --name mbuilder --use --bootstrap
                            ${buildxCmd}
                        """

                    }
                }
            }
        }

        stage('Build Publish Windows exe') {
            when {
                expression { return !buildSkipped }
            }
            steps {
                script {
                  sh "GOOS=windows GOARCH=amd64 go build -o netclient.exe main.go"
                }
                script {
                    def targetPath = "generic-local/netclient/${env.BRANCH_NAME}/${env.BUILD_NUMBER}/"
                    if (anyOf { patchIncrementBranchPatterns.collect { pattern ->
                        expression { env.BRANCH_NAME ==~ pattern }
                    } }) {
                        targetPath = "generic-local/netclient/${env.BRANCH_NAME}/${preIncVersion}/"
                    }
                }

                rtUpload (
                    serverId: 'dx-artifactory',
                    spec: """{
                            "files": [
                            {
                                "pattern": "netclient.exe",
                                "target": "${targetPath}"
                            }
                        ]
                    }"""
                )
            }
        }

        stage('Re-Commit Version Management') {
            when {
                expression { return !buildSkipped }
            }
            steps {
                script {
                    if (anyOf { patchIncrementBranchPatterns.collect { pattern ->
                            expression { env.BRANCH_NAME ==~ pattern }
                    } }) {
                        sh "git add VERSION"
                        sh "git commit -m '[skip ci] Update VERSION'"
                        sh "git push origin HEAD:${BRANCH_NAME}"
                    }

                }
            }
        }


    }
}