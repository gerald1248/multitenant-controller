package main

func apply(state map[string]string) error {
    policies, err := generatePolicies(state)
    if err != nil {
        return err;
	}
	
	for _, policy := range policies {
		err = applyPolicy(policy)
		if err != nil {
			return err
		}
	}

	err = prunePolicies(state)
	if err != nil {
		return err
	}

	return nil
}