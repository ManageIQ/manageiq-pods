package cr_migration

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
)

func migrate20210503163000(cr *miqv1alpha1.ManageIQ) *miqv1alpha1.ManageIQ {
	migrationId := "20210503163000"
	for _, migration := range cr.Spec.MigrationsRan {
		if migration == migrationId {
			return cr
		}
	}

	if cr.Spec.HttpdCpuLimit == "1000m" || cr.Spec.HttpdCpuLimit == "500m" {
		cr.Spec.HttpdCpuLimit = ""
	}

	if cr.Spec.HttpdCpuRequest == "50m" || cr.Spec.HttpdCpuRequest == "500m" {
		cr.Spec.HttpdCpuRequest = ""
	}

	if cr.Spec.HttpdMemoryLimit == "200Mi" || cr.Spec.HttpdMemoryLimit == "8192Mi" {
		cr.Spec.HttpdMemoryLimit = ""
	}

	if cr.Spec.HttpdMemoryRequest == "100Mi" || cr.Spec.HttpdMemoryRequest == "512Mi" {
		cr.Spec.HttpdMemoryRequest = ""
	}

	if cr.Spec.KafkaCpuLimit == "250m" {
		cr.Spec.KafkaCpuLimit = ""
	}

	if cr.Spec.KafkaCpuRequest == "250m" {
		cr.Spec.KafkaCpuRequest = ""
	}

	if cr.Spec.KafkaMemoryLimit == "1024Mi" {
		cr.Spec.KafkaMemoryLimit = ""
	}

	if cr.Spec.KafkaMemoryRequest == "256Mi" {
		cr.Spec.KafkaMemoryRequest = ""
	}

	if cr.Spec.MemcachedCpuLimit == "1000m" || cr.Spec.MemcachedCpuLimit == "200m" {
		cr.Spec.MemcachedCpuLimit = ""
	}

	if cr.Spec.MemcachedCpuRequest == "50m" || cr.Spec.MemcachedCpuRequest == "200m" {
		cr.Spec.MemcachedCpuRequest = ""
	}

	if cr.Spec.MemcachedMemoryLimit == "200Mi" || cr.Spec.MemcachedMemoryLimit == "256Mi" {
		cr.Spec.MemcachedMemoryLimit = ""
	}

	if cr.Spec.MemcachedMemoryRequest == "100Mi" || cr.Spec.MemcachedMemoryRequest == "64Mi" {
		cr.Spec.MemcachedMemoryRequest = ""
	}

	if cr.Spec.OrchestratorCpuLimit == "1000m" {
		cr.Spec.OrchestratorCpuLimit = ""
	}

	if cr.Spec.OrchestratorCpuRequest == "200m" || cr.Spec.OrchestratorCpuRequest == "1000m" {
		cr.Spec.OrchestratorCpuRequest = ""
	}

	if cr.Spec.OrchestratorMemoryLimit == "2048Mi" || cr.Spec.OrchestratorMemoryLimit == "1638Mi" {
		cr.Spec.OrchestratorMemoryLimit = ""
	}

	if cr.Spec.OrchestratorMemoryRequest == "1024Mi" || cr.Spec.OrchestratorMemoryRequest == "6144Mi" {
		cr.Spec.OrchestratorMemoryRequest = ""
	}

	if cr.Spec.PostgresqlCpuLimit == "1000m" || cr.Spec.PostgresqlCpuLimit == "500m" {
		cr.Spec.PostgresqlCpuLimit = ""
	}

	if cr.Spec.PostgresqlCpuRequest == "500m" {
		cr.Spec.PostgresqlCpuRequest = ""
	}

	if cr.Spec.PostgresqlMemoryLimit == "8192Mi" {
		cr.Spec.PostgresqlMemoryLimit = ""
	}

	if cr.Spec.PostgresqlMemoryRequest == "2048Mi" || cr.Spec.PostgresqlMemoryRequest == "8192Mi" {
		cr.Spec.PostgresqlMemoryRequest = ""
	}

	if cr.Spec.ZookeeperCpuLimit == "250m" {
		cr.Spec.ZookeeperCpuLimit = ""
	}

	if cr.Spec.ZookeeperCpuRequest == "250m" {
		cr.Spec.ZookeeperCpuRequest = ""
	}

	if cr.Spec.ZookeeperMemoryLimit == "256Mi" {
		cr.Spec.ZookeeperMemoryLimit = ""
	}

	if cr.Spec.ZookeeperMemoryRequest == "256Mi" {
		cr.Spec.ZookeeperMemoryRequest = ""
	}

	cr.Spec.MigrationsRan = append(cr.Spec.MigrationsRan, migrationId)

	return cr
}
