package edith

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Get started with Edith (requires root)",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if Ansible is installed, install dependencies and Ansible if not
		if !isAnsibleInstalled() {
			fmt.Println("Ansible is not installed. Install python3-pip and use it to install Ansible...")
			// installDependencies()
		}


		// Set UTF-8 locale
		setUTF8Locale()


		playbookFile := "playbook.yml"

		err := ioutil.WriteFile(playbookFile, []byte(ansiblePlaybook), 0644)
		if err != nil {
			fmt.Printf("Error writing playbook file: %v\n", err)
			os.Exit(1)
		}

		installCommand := exec.Command("ansible-playbook", playbookFile, "-i", "localhost,", "--connection", "local", "--extra-vars", fmt.Sprintf("ansible_user=%s", getCurrentUsername()))

		installCommand.Stdout = os.Stdout
		installCommand.Stderr = os.Stderr

		err = installCommand.Run()
		if err != nil {
			fmt.Printf("Error running Ansible playbook: %v\n", err)
			os.Exit(1)
		}

		err = os.Remove(playbookFile)
		if err != nil {
			fmt.Printf("Error removing playbook file: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func setUTF8Locale() {
	os.Setenv("LC_ALL", "C.UTF-8")
}

func isAnsibleInstalled() bool {
	cmd := exec.Command("ansible", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "ansible") || strings.Contains(string(output), "Ansible")
}

func installDependencies() {
	commands := []string{
		"sudo apt-get update",
		"sudo apt-get install -y python3-pip",
		"sudo python3 -m pip install ansible",
	}

	for _, command := range commands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error running command: %s, error: %v\n", command, err)
			os.Exit(1)
		}
	}
}

func getCurrentUsername() string {
	cmd := exec.Command("whoami")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting current username: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(output))
}


const ansiblePlaybook = `
- name: Setup Node
  hosts: all
  become: yes
  tasks:
    - name: Update packages
      apt:
        update_cache: yes
        cache_valid_time: 3600

    - name: Install required packages
      apt:
        name:
          - apt-transport-https
          - ca-certificates
          - curl
          - gnupg
          - lsb-release
          - build-essential
          - jq
          - htop
        state: present

    - name: Add Docker repository key
      ansible.builtin.apt_key:
        url: https://download.docker.com/linux/ubuntu/gpg
        state: present

    - name: Add Docker repository
      ansible.builtin.apt_repository:
        repo: "deb [arch=amd64] https://download.docker.com/linux/ubuntu {{ ansible_distribution_release }} stable"
        state: present

    - name: Install Docker
      apt:
        name: docker-ce
        state: present
        update_cache: yes

    - name: Add user to docker group
      user:
        name: "{{ ansible_user }}"
        groups: docker
        append: yes

    - name: Check if NVM is installed
      stat:
        path: "{{ ansible_env.HOME }}/.nvm/nvm.sh"
      register: nvm_installed 

    - name: Display nvm_installed
      debug:
        msg: "{{ nvm_installed }}"

    - name: Install NVM
      shell: |
        curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
      args:
        executable: /bin/bash
      become: no
      when: not nvm_installed.stat.exists

    - name: Install Node.js 16 and update npm 
      shell: |
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        nvm install 16.19.1
        nvm use v16.19.1
        npm i -g npm
      args:
        executable: /bin/bash
      become: no

    - name: Install Minikube
      block:
        - name: Install conntrack (required for Minikube)
          apt:
            name: conntrack
            state: present
          become: yes

        - name: Download Minikube binary
          get_url:
            url: https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
            dest: /tmp/minikube
            mode: '0755'

        - name: Install Minikube binary to /usr/local/bin
          command: sudo install /tmp/minikube /usr/local/bin/
          args:
            removes: /tmp/minikube
          become: yes

        - name: Check if kubectl command is available
          command: kubectl version --client
          register: kubectl_check_result
          ignore_errors: true
          become: 'no'

        - name: Set kubectl_available variable
          set_fact:
            kubectl_available: '{{ kubectl_check_result.rc == 0 }}'

        - name: Display kubectl_available variable
          debug:
            var: kubectl_available

        - name: Install kubectl
          block:
          - name: Get the latest stable kubectl version
            command: curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
            register: stable_kubectl_version
            become: no

          - name: Download kubectl binary
            get_url:
              url: "https://storage.googleapis.com/kubernetes-release/release/{{ stable_kubectl_version.stdout }}/bin/linux/amd64/kubectl"
              dest: /usr/local/bin/kubectl
              mode: '0755'
            become: yes

          - name: Display kubectl version
            debug:
              msg: "{{ stable_kubectl_version.stdout }}"
        
        - name: Download pachctl_1.12.5_amd64.deb
          get_url:
            url: https://github.com/pachyderm/pachyderm/releases/download/v1.12.5/pachctl_1.12.5_amd64.deb
            dest: /tmp/pachctl_1.12.5_amd64.deb
            mode: 0644

        - name: Install pachctl
          ansible.builtin.apt:
            deb: /tmp/pachctl_1.12.5_amd64.deb

        - name: Ensure kubefwd Docker image is present
          docker_image:
            name: "txn2/kubefwd"
            source: pull
    
        - name: Run kubefwd container
          docker_container:
            name: default
            image: txn2/kubefwd
            command: services -n default
            state: started
            recreate: yes
            privileged: yes
            interactive: yes
            tty: yes
            volumes:
              - "{{ lookup('env', 'HOME') }}/.kube/:/root/.kube/"
          register: kubefwd_container
    
        - name: Display kubefwd container information
          debug:
            var: kubefwd_container

        - name: Create Minikube service file
          copy:
            content: |
              [Unit]
              Description=Minikube
              After=network.target

              [Service]
              Type=oneshot
              RemainAfterExit=yes
              Environment="MINIKUBE_HOME=/home/{{ ansible_user }}"
              ExecStart=/usr/local/bin/minikube start --wait=all --cpus=2 --memory=2GB --force-systemd=true
              ExecStop=/usr/local/bin/minikube stop

              [Install]
              WantedBy=multi-user.target
            dest: /etc/systemd/system/minikube.service
            owner: root
            group: root
            mode: 0644

        - name: Reload systemd daemon
          systemd:
            daemon_reload: yes

        - name: Enable and start Minikube service
          systemd:
            name: minikube.service
            enabled: yes
            state: started
`