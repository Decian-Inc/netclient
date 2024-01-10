/* groovylint-disable LineLength */
def majorVersion = ''
def minorVersion = ''
def patchVersion = ''
def dx_patchVersion = ''
def buildSkipped = false


pipeline {
    agent {
        label 'ubuntu22-vm'
    }
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
                    def version = readFile("${env.WORKSPACE}/VERSION").trim()
                    (majorVersion, minorVersion, patchVersion, dx_patchVersion) = version.tokenize('.')

                   // display version info
                    echo "Current Version: ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"

                    if (env.BRANCH_NAME == 'main' && !buildSkipped) {
                        // Bump Patch Version, commit
                        patchVersion = patchVersion.toInteger() + 1
                        echo "New Version: ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"
                        sh "echo ${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion} > VERSION"
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
                    def version = "${majorVersion}.${minorVersion}.${patchVersion}.${dx_patchVersion}"
                    def dockerTags = []
                    dockerTags.add("${version}-${env.BRANCH_NAME.replaceAll("/", "-")}-${env.BUILD_NUMBER}")
                    dockerTags.add("${version}-${env.BRANCH_NAME.replaceAll("/", "-")}")

                    if (env.BRANCH_NAME == 'main') {
                        dockerTags.add("${version}")
                        dockerTags.add("${majorVersion}.${minorVersion}.${patchVersion}")
                        dockerTags.add("${majorVersion}.${minorVersion}")
                        dockerTags.add("${majorVersion}")
                    }


                    def dockerBuildCommandTags = dockerTags.collect { tag -> "-t $DOCKER_REGISTRY/$IMAGE_NAME:${tag}" }.join(' ')



                    docker.withRegistry('https://nexus-registry.decian.net', 'nexus-docker-writer-username-password') {
                          // Build and push the image
                          // sh """
                          //   docker build --build-arg VERSION=$version --push $dockerBuildCommandTags .
                          // """
                          def buildCommand = "docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 --push $dockerBuildCommandTags ."
                          sh "echo \"${buildCommand}\""

                          sh """
                              docker buildx create --name mbuilder --use --bootstrap
                              ${buildCommand}
                          """

                    }
                }
            }
        }

        stage('Re-Commit Version Management') {
            when {
                expression { return !buildSkipped }
            }
            steps {
                script {
                    if (env.BRANCH_NAME == 'main') {
                        sh "git add VERSION"
                        sh "git commit -m '[skip ci] Update VERSION'"
                        sh "git push origin HEAD:main"
                    }

                }
            }
        }


    }
}
