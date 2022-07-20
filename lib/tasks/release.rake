namespace :release do
  desc "Tasks to run on a new branch when a new branch is created"
  task :new_branch do
    require 'pathname'

    branch = ENV["RELEASE_BRANCH"]
    if branch.nil? || branch.empty?
      STDERR.puts "ERROR: You must set the env var RELEASE_BRANCH to the proper value."
      exit 1
    end

    current_branch = `git rev-parse --abbrev-ref HEAD`.chomp
    if current_branch == "master"
      STDERR.puts "ERROR: You cannot do new branch tasks from the master branch."
      exit 1
    end

    root = Pathname.new(__dir__).join("../..")

    branch_number = branch[0].ord - 96
    rpm_repo_name = "#{branch_number}-#{branch}"

    # Modify README
    readme = root.join("README.md")
    content = readme.read
    readme.write(content.sub(/^(default: )\w+/, "\\1latest-#{branch}"))

    # Modify operator README
    operator_readme = root.join("manageiq-operator", "README.md")
    content = operator_readme.read
    operator_readme.write(content.gsub(%r{(/manageiq-operator:)[-\w]+}, "\\1latest-#{branch}"))

    # Modify CR
    cr = root.join("manageiq-operator", "helpers", "miq-components", "cr.go")
    content = cr.read
    cr.write(content.sub(/(cr\.Spec\.OrchestratorImageTag.+?\n\s+return ")[^"]+(")/, "\\1latest-#{branch}\\2"))

    # Modify types
    types = root.join("manageiq-operator", "api", "v1alpha1", "manageiq_types.go")
    content = types.read
    types.write(content.sub(/(tag used for the orchestrator and worker deployments \(default: )[^\)]+(\))/, "\\1latest-#{branch}\\2"))

    # Modify operator deployment yaml
    deploy_operator = root.join("config", "manager", "manager.yaml")
    content = deploy_operator.read
    deploy_operator.write(content.sub(%r{(docker.io/manageiq/manageiq-operator:).+$}, "\\1latest-#{branch}"))

    # Modify deploy CRD
    deploy_crd = root.join("manageiq-operator", "crd", "bases", "manageiq.org_manageiqs.yaml")
    content = deploy_crd.read
    deploy_crd.write(content.sub(/(tag used for the orchestrator and worker deployments\n\s+\(default: )[^\)]+(\))/, "\\1latest-#{branch}\\2"))

    # Modify bin/build
    build_script = root.join("bin", "build")
    content = build_script.read
    content.sub!(/^(TAG=).+$/, "\\1latest-#{branch}")
    content.sub!(/(BUILD_REF:-)\w+(\})/, "\\1#{branch}\\2")
    build_script.write(content)

    # Modify bin/remove_images
    remove_script = root.join("bin", "remove_images")
    content = remove_script.read
    remove_script.write(content.sub(/^(TAG=).+$/, "\\1latest-#{branch}"))

    # Modify base Dockerfile
    base_dockerfile = root.join("images", "manageiq-base", "Dockerfile")
    content = base_dockerfile.read
    content.sub!(/^(ARG BUILD_REF=)\w+/, "\\1#{branch}")
    content.sub!(%r{(/rpm.manageiq.org/release/)\d+-\w+}, "\\1#{rpm_repo_name}")
    content.sub!(%r{(/el8/noarch/manageiq-release-)\d+\.\d+-\d+}, "\\1#{branch_number}.0-1")
    content.sub!(/(manageiq-)\d+-\w+(-nightly)/, "\\1#{rpm_repo_name}\\2")
    base_dockerfile.write(content)

    # Modify Dockerfiles
    dockerfiles = %w[manageiq-base-worker manageiq-webserver-worker manageiq-ui-worker manageiq-orchestrator].map do |worker|
      root.join("images", worker, "Dockerfile").tap do |dockerfile|
        content = dockerfile.read
        dockerfile.write(content.sub(/^(ARG FROM_TAG=).+$/, "\\1latest-#{branch}"))
      end
    end

    # Commit
    files_to_update = [readme, operator_readme, cr, types, deploy_operator, deploy_crd, deploy_csv, catalog_crd, build_script, remove_script, base_dockerfile, *dockerfiles]
    exit $?.exitstatus unless system("git add #{files_to_update.join(" ")}")
    exit $?.exitstatus unless system("git commit -m 'Changes for new branch #{branch}'")

    puts
    puts "The commit on #{current_branch} has been created."
    puts "Run the following to push to the upstream remote:"
    puts
    puts "\tgit push upstream #{current_branch}"
    puts
  end

  desc "Tasks to run on the master branch when a new branch is created"
  task :new_branch_master do
    require 'pathname'

    branch = ENV["RELEASE_BRANCH"]
    if branch.nil? || branch.empty?
      STDERR.puts "ERROR: You must set the env var RELEASE_BRANCH to the proper value."
      exit 1
    end

    next_branch = ENV["RELEASE_BRANCH_NEXT"]
    if next_branch.nil? || next_branch.empty?
      STDERR.puts "ERROR: You must set the env var RELEASE_BRANCH_NEXT to the proper value."
      exit 1
    end

    current_branch = `git rev-parse --abbrev-ref HEAD`.chomp
    if current_branch != "master"
      STDERR.puts "ERROR: You cannot do master branch tasks from a non-master branch (#{current_branch})."
      exit 1
    end

    root = Pathname.new(__dir__).join("../..")

    next_branch_number = next_branch[0].ord - 96
    rpm_repo_name = "#{next_branch_number}-#{next_branch}"

    # Modify base Dockerfile
    base_dockerfile = root.join("images", "manageiq-base", "Dockerfile")
    content = base_dockerfile.read
    content.sub!(%r{(/rpm.manageiq.org/release/)\d+-\w+}, "\\1#{rpm_repo_name}")
    content.sub!(%r{(/el8/noarch/manageiq-release-)\d+\.\d+-\d+}, "\\1#{next_branch_number}.0-1")
    content.sub!(/(manageiq-)\d+-\w+(-nightly)/, "\\1#{rpm_repo_name}\\2")
    base_dockerfile.write(content)

    # Commit
    files_to_update = [base_dockerfile]
    exit $?.exitstatus unless system("git add #{files_to_update.join(" ")}")
    exit $?.exitstatus unless system("git commit -m 'Changes after new branch #{branch}'")

    puts
    puts "The commit on #{current_branch} has been created."
    puts "Run the following to push to the upstream remote:"
    puts
    puts "\tgit push upstream #{current_branch}"
    puts
  end
end
